package eks

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	eksv1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	"github.com/rancher/eks-operator/pkg/eks/services"
	"github.com/rancher/eks-operator/templates"
	"github.com/rancher/eks-operator/utils"
)

const (
	// CloudFormation stack statuses
	createInProgressStatus   = "CREATE_IN_PROGRESS"
	createCompleteStatus     = "CREATE_COMPLETE"
	createFailedStatus       = "CREATE_FAILED"
	rollbackInProgressStatus = "ROLLBACK_IN_PROGRESS"

	LaunchTemplateNameFormat = "rancher-managed-lt-%s"
	launchTemplateTagKey     = "rancher-managed-template"
	launchTemplateTagValue   = "do-not-modify-or-delete"
	defaultStorageDeviceName = "/dev/xvda"

	defaultAudienceOpenIDConnect = "sts.amazonaws.com"
	ebsCSIAddonName              = "aws-ebs-csi-driver"
)

type CreateClusterOptions struct {
	EKSService services.EKSServiceInterface
	Config     *eksv1.EKSClusterConfig
	RoleARN    string
}

func CreateCluster(opts CreateClusterOptions) (*eks.CreateClusterOutput, error) {
	createClusterInput := newClusterInput(opts.Config, opts.RoleARN)

	return opts.EKSService.CreateCluster(createClusterInput)
}

func newClusterInput(config *eksv1.EKSClusterConfig, roleARN string) *eks.CreateClusterInput {
	createClusterInput := &eks.CreateClusterInput{
		Name:    aws.String(config.Spec.DisplayName),
		RoleArn: aws.String(roleARN),
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			EndpointPrivateAccess: config.Spec.PrivateAccess,
			EndpointPublicAccess:  config.Spec.PublicAccess,
			SecurityGroupIds:      aws.StringSlice(config.Status.SecurityGroups),
			SubnetIds:             aws.StringSlice(config.Status.Subnets),
			PublicAccessCidrs:     getPublicAccessCidrs(config.Spec.PublicAccessSources),
		},
		Tags:    getTags(config.Spec.Tags),
		Logging: getLogging(config.Spec.LoggingTypes),
		Version: config.Spec.KubernetesVersion,
	}

	if aws.BoolValue(config.Spec.SecretsEncryption) {
		createClusterInput.EncryptionConfig = []*eks.EncryptionConfig{
			{
				Provider: &eks.Provider{
					KeyArn: config.Spec.KmsKey,
				},
				Resources: aws.StringSlice([]string{"secrets"}),
			},
		}
	}

	return createClusterInput
}

type CreateStackOptions struct {
	CloudFormationService services.CloudFormationServiceInterface
	StackName             string
	DisplayName           string
	TemplateBody          string
	Capabilities          []string
	Parameters            []*cloudformation.Parameter
}

func CreateStack(opts CreateStackOptions) (*cloudformation.DescribeStacksOutput, error) {
	_, err := opts.CloudFormationService.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String(opts.StackName),
		TemplateBody: aws.String(opts.TemplateBody),
		Capabilities: aws.StringSlice(opts.Capabilities),
		Parameters:   opts.Parameters,
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String("displayName"),
				Value: aws.String(opts.DisplayName),
			},
		},
	})
	if err != nil && !alreadyExistsInCloudFormationError(err) {
		return nil, fmt.Errorf("error creating master: %v", err)
	}

	var stack *cloudformation.DescribeStacksOutput
	status := createInProgressStatus

	for status == createInProgressStatus {
		time.Sleep(time.Second * 5)
		stack, err = opts.CloudFormationService.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: aws.String(opts.StackName),
		})
		if err != nil {
			return nil, fmt.Errorf("error polling stack info: %v", err)
		}

		if stack == nil || stack.Stacks == nil || len(stack.Stacks) == 0 {
			return nil, fmt.Errorf("stack did not have output: %v", err)
		}

		status = *stack.Stacks[0].StackStatus
	}

	if status != createCompleteStatus {
		reason := "reason unknown"
		events, err := opts.CloudFormationService.DescribeStackEvents(&cloudformation.DescribeStackEventsInput{
			StackName: aws.String(opts.StackName),
		})
		if err == nil {
			for _, event := range events.StackEvents {
				// guard against nil pointer dereference
				if event.ResourceStatus == nil || event.LogicalResourceId == nil || event.ResourceStatusReason == nil {
					continue
				}

				if *event.ResourceStatus == createFailedStatus {
					reason = *event.ResourceStatusReason
					break
				}

				if *event.ResourceStatus == rollbackInProgressStatus {
					reason = *event.ResourceStatusReason
					// do not break so that CREATE_FAILED takes priority
				}
			}
		}
		return nil, fmt.Errorf("stack failed to create: %v", reason)
	}

	return stack, nil
}

type CreateLaunchTemplateOptions struct {
	EC2Service services.EC2ServiceInterface
	Config     *eksv1.EKSClusterConfig
}

func CreateLaunchTemplate(opts CreateLaunchTemplateOptions) error {
	_, err := opts.EC2Service.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateIds: []*string{aws.String(opts.Config.Status.ManagedLaunchTemplateID)},
	})
	if opts.Config.Status.ManagedLaunchTemplateID == "" || doesNotExist(err) {
		lt, err := createLaunchTemplate(opts.EC2Service, opts.Config.Spec.DisplayName)
		if err != nil {
			return fmt.Errorf("error creating launch template: %w", err)
		}
		opts.Config.Status.ManagedLaunchTemplateID = aws.StringValue(lt.ID)
	} else if err != nil {
		return fmt.Errorf("error checking for existing launch template: %w", err)
	}

	return nil
}

func createLaunchTemplate(ec2Service services.EC2ServiceInterface, clusterDisplayName string) (*eksv1.LaunchTemplate, error) {
	// The first version of the rancher-managed launch template will be the default version.
	// Since the default version cannot be deleted until the launch template is deleted, it will not be used for any node group.
	// Also, launch templates cannot be created blank, so fake userdata is added to the first version.
	launchTemplateCreateInput := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateData: &ec2.RequestLaunchTemplateData{UserData: aws.String("cGxhY2Vob2xkZXIK")},
		LaunchTemplateName: aws.String(fmt.Sprintf(LaunchTemplateNameFormat, clusterDisplayName)),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(ec2.ResourceTypeLaunchTemplate),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String(launchTemplateTagKey),
						Value: aws.String(launchTemplateTagValue),
					},
				},
			},
		},
	}

	awsLaunchTemplateOutput, err := ec2Service.CreateLaunchTemplate(launchTemplateCreateInput)
	if err != nil {
		return nil, err
	}

	return &eksv1.LaunchTemplate{
		Name:    awsLaunchTemplateOutput.LaunchTemplate.LaunchTemplateName,
		ID:      awsLaunchTemplateOutput.LaunchTemplate.LaunchTemplateId,
		Version: awsLaunchTemplateOutput.LaunchTemplate.LatestVersionNumber,
	}, nil
}

type CreateNodeGroupOptions struct {
	EC2Service            services.EC2ServiceInterface
	CloudFormationService services.CloudFormationServiceInterface
	EKSService            services.EKSServiceInterface

	Config    *eksv1.EKSClusterConfig
	NodeGroup eksv1.NodeGroup
}

func CreateNodeGroup(opts CreateNodeGroupOptions) (string, string, error) {
	var err error
	capacityType := eks.CapacityTypesOnDemand
	if aws.BoolValue(opts.NodeGroup.RequestSpotInstances) {
		capacityType = eks.CapacityTypesSpot
	}
	nodeGroupCreateInput := &eks.CreateNodegroupInput{
		ClusterName:   aws.String(opts.Config.Spec.DisplayName),
		NodegroupName: opts.NodeGroup.NodegroupName,
		Labels:        opts.NodeGroup.Labels,
		ScalingConfig: &eks.NodegroupScalingConfig{
			DesiredSize: opts.NodeGroup.DesiredSize,
			MaxSize:     opts.NodeGroup.MaxSize,
			MinSize:     opts.NodeGroup.MinSize,
		},
		CapacityType: aws.String(capacityType),
	}

	lt := opts.NodeGroup.LaunchTemplate

	if lt == nil {
		// In this case, the user has not specified their own launch template.
		// If the cluster doesn't have a launch template associated with it, then we create one.
		lt, err = CreateNewLaunchTemplateVersion(opts.EC2Service, opts.Config.Status.ManagedLaunchTemplateID, opts.NodeGroup)
		if err != nil {
			return "", "", err
		}
	}

	var launchTemplateVersion *string
	if aws.Int64Value(lt.Version) != 0 {
		launchTemplateVersion = aws.String(strconv.FormatInt(*lt.Version, 10))
	}

	nodeGroupCreateInput.LaunchTemplate = &eks.LaunchTemplateSpecification{
		Id:      lt.ID,
		Version: launchTemplateVersion,
	}

	if aws.BoolValue(opts.NodeGroup.RequestSpotInstances) {
		nodeGroupCreateInput.InstanceTypes = opts.NodeGroup.SpotInstanceTypes
	}

	if aws.StringValue(opts.NodeGroup.ImageID) == "" {
		if gpu := opts.NodeGroup.Gpu; aws.BoolValue(gpu) {
			nodeGroupCreateInput.AmiType = aws.String(eks.AMITypesAl2X8664Gpu)
		} else {
			nodeGroupCreateInput.AmiType = aws.String(eks.AMITypesAl2X8664)
		}
	}

	if len(opts.NodeGroup.Subnets) != 0 {
		nodeGroupCreateInput.Subnets = aws.StringSlice(opts.NodeGroup.Subnets)
	} else {
		nodeGroupCreateInput.Subnets = aws.StringSlice(opts.Config.Status.Subnets)
	}

	generatedNodeRole := opts.Config.Status.GeneratedNodeRole

	if aws.StringValue(opts.NodeGroup.NodeRole) == "" {
		if opts.Config.Status.GeneratedNodeRole == "" {
			finalTemplate := fmt.Sprintf(templates.NodeInstanceRoleTemplate, getEC2ServiceEndpoint(opts.Config.Spec.Region))
			output, err := CreateStack(CreateStackOptions{
				CloudFormationService: opts.CloudFormationService,
				StackName:             fmt.Sprintf("%s-node-instance-role", opts.Config.Spec.DisplayName),
				DisplayName:           opts.Config.Spec.DisplayName,
				TemplateBody:          finalTemplate,
				Capabilities:          []string{cloudformation.CapabilityCapabilityIam},
				Parameters:            []*cloudformation.Parameter{},
			})
			if err != nil {
				// If there was an error creating the node role stack, return an empty launch template
				// version and the error.
				return "", "", err
			}
			generatedNodeRole = getParameterValueFromOutput("NodeInstanceRole", output.Stacks[0].Outputs)
		}
		nodeGroupCreateInput.NodeRole = aws.String(generatedNodeRole)
	} else {
		nodeGroupCreateInput.NodeRole = opts.NodeGroup.NodeRole
	}

	_, err = opts.EKSService.CreateNodegroup(nodeGroupCreateInput)
	if err != nil {
		// If there was an error creating the node group, then the template version should be deleted
		// to prevent many launch template versions from being created before the issue is fixed.
		DeleteLaunchTemplateVersions(opts.EC2Service, *lt.ID, []*string{launchTemplateVersion})
	}

	// Return the launch template version and generated node role to the calling function so they can
	// be set on the Status.
	return aws.StringValue(launchTemplateVersion), generatedNodeRole, err
}

func CreateNewLaunchTemplateVersion(
	ec2Service services.EC2ServiceInterface,
	launchTemplateID string,
	group eksv1.NodeGroup,
) (*eksv1.LaunchTemplate, error) {
	launchTemplate, err := buildLaunchTemplateData(ec2Service, group)
	if err != nil {
		return nil, err
	}

	launchTemplateVersionInput := &ec2.CreateLaunchTemplateVersionInput{
		LaunchTemplateData: launchTemplate,
		LaunchTemplateId:   aws.String(launchTemplateID),
	}

	awsLaunchTemplateOutput, err := ec2Service.CreateLaunchTemplateVersion(launchTemplateVersionInput)
	if err != nil {
		return nil, err
	}

	return &eksv1.LaunchTemplate{
		Name:    awsLaunchTemplateOutput.LaunchTemplateVersion.LaunchTemplateName,
		ID:      awsLaunchTemplateOutput.LaunchTemplateVersion.LaunchTemplateId,
		Version: awsLaunchTemplateOutput.LaunchTemplateVersion.VersionNumber,
	}, nil
}

func buildLaunchTemplateData(
	ec2Service services.EC2ServiceInterface,
	group eksv1.NodeGroup,
) (*ec2.RequestLaunchTemplateData, error) {
	var imageID *string
	if aws.StringValue(group.ImageID) != "" {
		imageID = group.ImageID
	}

	userdata := group.UserData
	if aws.StringValue(userdata) != "" {
		if !strings.Contains(*userdata, "Content-Type: multipart/mixed") {
			return nil, fmt.Errorf(
				"userdata for nodegroup [%s] is not of mime time multipart/mixed",
				aws.StringValue(group.NodegroupName),
			)
		}
		*userdata = base64.StdEncoding.EncodeToString([]byte(*userdata))
	}

	deviceName := aws.String(defaultStorageDeviceName)
	if aws.StringValue(group.ImageID) != "" {
		if rootDeviceName, err := getImageRootDeviceName(ec2Service, group.ImageID); err != nil {
			return nil, err
		} else if rootDeviceName != nil {
			deviceName = rootDeviceName
		}
	}

	launchTemplateData := &ec2.RequestLaunchTemplateData{
		ImageId:  imageID,
		KeyName:  group.Ec2SshKey,
		UserData: userdata,
		BlockDeviceMappings: []*ec2.LaunchTemplateBlockDeviceMappingRequest{
			{
				DeviceName: deviceName,
				Ebs: &ec2.LaunchTemplateEbsBlockDeviceRequest{
					VolumeSize: group.DiskSize,
				},
			},
		},
		TagSpecifications: utils.CreateTagSpecs(group.ResourceTags),
	}
	if !aws.BoolValue(group.RequestSpotInstances) {
		launchTemplateData.InstanceType = group.InstanceType
	}

	return launchTemplateData, nil
}

func getImageRootDeviceName(ec2Service services.EC2ServiceInterface, imageID *string) (*string, error) {
	if imageID == nil {
		return nil, fmt.Errorf("imageID is nil")
	}
	describeOutput, err := ec2Service.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{imageID}})
	if err != nil {
		return nil, err
	}
	if len(describeOutput.Images) == 0 {
		return nil, fmt.Errorf("no images returned for id %v", aws.StringValue(imageID))
	}

	return describeOutput.Images[0].RootDeviceName, nil
}

func getTags(tags map[string]string) map[string]*string {
	if len(tags) == 0 {
		return nil
	}

	return aws.StringMap(tags)
}

func getLogging(loggingTypes []string) *eks.Logging {
	if len(loggingTypes) == 0 {
		return &eks.Logging{
			ClusterLogging: []*eks.LogSetup{
				{
					Enabled: aws.Bool(false),
					Types:   aws.StringSlice(loggingTypes),
				},
			},
		}
	}
	return &eks.Logging{
		ClusterLogging: []*eks.LogSetup{
			{
				Enabled: aws.Bool(true),
				Types:   aws.StringSlice(loggingTypes),
			},
		},
	}
}

func getPublicAccessCidrs(publicAccessCidrs []string) []*string {
	if len(publicAccessCidrs) == 0 {
		return aws.StringSlice([]string{"0.0.0.0/0"})
	}

	return aws.StringSlice(publicAccessCidrs)
}

func alreadyExistsInCloudFormationError(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case cloudformation.ErrCodeAlreadyExistsException:
			return true
		}
	}

	return false
}

func doesNotExist(err error) bool {
	// There is no better way of doing this because AWS API does not distinguish between a attempt to delete a stack
	// (or key pair) that does not exist, and, for example, a malformed delete request, so we have to parse the error
	// message
	if err != nil {
		return strings.Contains(err.Error(), "does not exist")
	}

	return false
}

func getEC2ServiceEndpoint(region string) string {
	if p, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), region); ok {
		return fmt.Sprintf("%s.%s", ec2.ServiceName, p.DNSSuffix())
	}
	return "ec2.amazonaws.com"
}

func getParameterValueFromOutput(key string, outputs []*cloudformation.Output) string {
	for _, output := range outputs {
		if *output.OutputKey == key {
			return *output.OutputValue
		}
	}

	return ""
}

// EnableEBSCSIDriverInput holds the options for installing the EBS CSI driver
type EnableEBSCSIDriverInput struct {
	EKSService     services.EKSServiceInterface
	IAMService     services.IAMServiceInterface
	CFService      services.CloudFormationServiceInterface
	Config         *eksv1.EKSClusterConfig
	OIDCProviderID string
	DriverRoleARN  string
}

// EnableEBSCSIDriver manages the EBS CSI driver installation, including the creation of the OIDC Provider,
// the IAM role and the validation and installation of the EKS add-on
func EnableEBSCSIDriver(opts EnableEBSCSIDriverInput) {}

// ConfigureOIDCProvider creates a new Open ID Connect Provider associated with the cluster
// if there are no providers available
// func ConfigureOIDCProvider(config *eksv1.EKSClusterConfig, iamService services.IAMServiceInterface, eksService services.EKSServiceInterface) error {
func ConfigureOIDCProvider(opts EnableEBSCSIDriverInput) error {
	output, err := opts.IAMService.ListOIDCProviders(&iam.ListOpenIDConnectProvidersInput{})
	if err != nil {
		return fmt.Errorf("error listing oidc providers: %v", err)
	}
	clusterOutput, err := opts.EKSService.DescribeCluster(&eks.DescribeClusterInput{
		Name: aws.String(opts.Config.Spec.DisplayName),
	})
	id := path.Base(*clusterOutput.Cluster.Identity.Oidc.Issuer)

	for _, prov := range output.OpenIDConnectProviderList {
		if strings.Contains(*prov.Arn, id) {
			// TODO: review this and how to proceed with creation of OIDC provider
			// how to pass the OIDC provider ID to the EBS CSI driver?
			opts.OIDCProviderID = path.Base(*prov.Arn)
			return nil
		}
	}

	thumbprint, err := getIssuerThumbprint(*clusterOutput.Cluster.Identity.Oidc.Issuer)
	if err != nil {
		return fmt.Errorf("error getting server certificate tumbprints for OIDC: %v", err)
	}
	input := &iam.CreateOpenIDConnectProviderInput{
		ClientIDList:   []*string{aws.String(defaultAudienceOpenIDConnect)},
		ThumbprintList: []*string{&thumbprint},
		Url:            clusterOutput.Cluster.Identity.Oidc.Issuer,
		Tags:           []*iam.Tag{},
	}
	_, err = opts.IAMService.CreateOIDCProvider(input)
	if err != nil {
		return fmt.Errorf("creating OIDC provider: %v", err)
	}

	return nil
}

func getIssuerThumbprint(issuer string) (string, error) {
	issuerURL, err := url.Parse(issuer)
	if err != nil {
		return "", fmt.Errorf("parsing issuer url: %w", err)
	}
	if issuerURL.Port() == "" {
		issuerURL.Host += ":443"
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				MinVersion:         tls.VersionTLS12,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Get(issuerURL.String())
	if err != nil {
		return "", fmt.Errorf("querying oidc issuer endpoint %s: %w", issuerURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return "", errors.New("unable to get oidc issuers cert")
	}

	root := resp.TLS.PeerCertificates[len(resp.TLS.PeerCertificates)-1]

	return fmt.Sprintf("%x", sha1.Sum(root.Raw)), nil
}

// CreateEBSCSIDriverRole creates an IAM role for the EKS cluster to interact with
// EBS through the previously created Open ID Connect provider
func CreateEBSCSIDriverRole(opts EnableEBSCSIDriverInput) (string, error) {
	templateData := struct {
		Region     string
		ProviderID string
	}{
		Region:     opts.Config.Spec.Region,
		ProviderID: opts.OIDCProviderID,
	}
	tmpl, err := template.New("ebsrole").Parse(templates.EBSCSIDriverTemplate)
	if err != nil {
		return "", fmt.Errorf("parsing ebs role template: %v", err)
	}
	buf := &bytes.Buffer{}
	if execErr := tmpl.Execute(buf, templateData); execErr != nil {
		return "", fmt.Errorf("executing ebs role template: %v", err)
	}
	finalTemplate := buf.String()

	output, err := CreateStack(CreateStackOptions{
		CloudFormationService: opts.CFService,
		StackName:             fmt.Sprintf("%s-ebs-csi-driver-role", opts.Config.Spec.DisplayName),
		DisplayName:           opts.Config.Spec.DisplayName,
		TemplateBody:          finalTemplate,
		Capabilities:          []string{cloudformation.CapabilityCapabilityIam},
		Parameters:            []*cloudformation.Parameter{},
	})
	if err != nil {
		// If there was an error creating the driver role stack, return an empty role arn and the error
		return "", fmt.Errorf("creating ebs csi driver role: %v", err)
	}
	createdRoleArn := getParameterValueFromOutput("EBSCSIDriverRole", output.Stacks[0].Outputs)

	return createdRoleArn, nil
}

func checkEBSAddon(opts EnableEBSCSIDriverInput) (string, error) {
	input := eks.DescribeAddonInput{
		AddonName:   aws.String(ebsCSIAddonName),
		ClusterName: aws.String(opts.Config.Spec.DisplayName),
	}

	output, err := opts.EKSService.DescribeAddon(&input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == eks.ErrCodeResourceNotFoundException {
				log.Println("EBS CSI driver addon not found: got resource not found exception")
				return "", nil
			}
		}
	}
	if output == nil {
		log.Println("EBS CSI driver addon not found")
		return "", nil
	}
	log.Println("EBS CSI driver addon found:", *output.Addon.AddonArn)

	return *output.Addon.AddonArn, nil
}

func installEBSAddon(opts EnableEBSCSIDriverInput) error {
	input := eks.CreateAddonInput{
		AddonName:             aws.String(ebsCSIAddonName),
		ClusterName:           aws.String(opts.Config.Spec.DisplayName),
		ServiceAccountRoleArn: aws.String(opts.DriverRoleARN),
	}

	output, err := opts.EKSService.CreateAddon(&input)
	if err != nil {
		return fmt.Errorf("cannot install EBS CSI driver addon: %v", err)
	}
	fmt.Println("installed addon EBS CSI driver:", *output.Addon.AddonArn)

	return nil
}
