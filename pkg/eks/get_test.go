package eks

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	eksv1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	"github.com/rancher/eks-operator/pkg/eks/services/mock_services"
)

var _ = Describe("GetClusterState", func() {
	var (
		mockController          *gomock.Controller
		eksServiceMock          *mock_services.MockEKSServiceInterface
		getClusterStatusOptions *GetClusterStatusOpts
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		eksServiceMock = mock_services.NewMockEKSServiceInterface(mockController)
		getClusterStatusOptions = &GetClusterStatusOpts{
			EKSService: eksServiceMock,
			Config: &eksv1.EKSClusterConfig{
				Spec: eksv1.EKSClusterConfigSpec{
					DisplayName: "test-cluster",
				},
			},
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	It("should successfully get cluster state", func() {
		eksServiceMock.EXPECT().DescribeCluster(
			&eks.DescribeClusterInput{
				Name: aws.String(getClusterStatusOptions.Config.Spec.DisplayName),
			},
		).Return(&eks.DescribeClusterOutput{}, nil)
		clusterState, err := GetClusterState(getClusterStatusOptions)
		Expect(err).ToNot(HaveOccurred())
		Expect(clusterState).ToNot(BeNil())
	})

	It("should fail to get cluster state", func() {
		eksServiceMock.EXPECT().DescribeCluster(gomock.Any()).Return(nil, errors.New("error getting cluster state"))
		_, err := GetClusterState(getClusterStatusOptions)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("GetLaunchTemplateVersions", func() {
	var (
		mockController           *gomock.Controller
		ec2ServiceMock           *mock_services.MockEC2ServiceInterface
		getLaunchTemplateOptions *GetLaunchTemplateVersionsOpts
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		ec2ServiceMock = mock_services.NewMockEC2ServiceInterface(mockController)
		getLaunchTemplateOptions = &GetLaunchTemplateVersionsOpts{
			EC2Service:       ec2ServiceMock,
			LaunchTemplateID: aws.String("test-launch-template-id"),
			Versions:         aws.StringSlice([]string{"1", "2"}),
		}
	})

	AfterEach(func() {
		mockController.Finish()
	})

	It("should successfully get launch template versions", func() {
		ec2ServiceMock.EXPECT().DescribeLaunchTemplateVersions(
			&ec2.DescribeLaunchTemplateVersionsInput{
				LaunchTemplateId: getLaunchTemplateOptions.LaunchTemplateID,
				Versions:         getLaunchTemplateOptions.Versions,
			},
		).Return(&ec2.DescribeLaunchTemplateVersionsOutput{}, nil)
		ltVersion, err := GetLaunchTemplateVersions(getLaunchTemplateOptions)
		Expect(err).ToNot(HaveOccurred())
		Expect(ltVersion).ToNot(BeNil())
	})

	It("should fail to get launch template versions", func() {
		ec2ServiceMock.EXPECT().DescribeLaunchTemplateVersions(gomock.Any()).Return(nil, errors.New("error getting launch template versions"))
		_, err := GetLaunchTemplateVersions(getLaunchTemplateOptions)
		Expect(err).To(HaveOccurred())
	})

	It("should fail to get launch template versions when template id is missing", func() {
		getLaunchTemplateOptions.LaunchTemplateID = nil
		_, err := GetLaunchTemplateVersions(getLaunchTemplateOptions)
		Expect(err).To(HaveOccurred())
	})

	It("should fail to get launch template versions when versions are missing", func() {
		getLaunchTemplateOptions.Versions = nil
		_, err := GetLaunchTemplateVersions(getLaunchTemplateOptions)
		Expect(err).To(HaveOccurred())
	})

	var _ = Describe("GetEBSCSIAddon", func() {
		var (
			mockController            *gomock.Controller
			eksServiceMock            *mock_services.MockEKSServiceInterface
			iamServiceMock            *mock_services.MockIAMServiceInterface
			cloudFormationServiceMock *mock_services.MockCloudFormationServiceInterface
			eksDescribeAddonOutput    *eks.DescribeAddonOutput
			enableEBSCSIDriverInput   *EnableEBSCSIDriverInput
		)

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			eksServiceMock = mock_services.NewMockEKSServiceInterface(mockController)
			iamServiceMock = mock_services.NewMockIAMServiceInterface(mockController)
			cloudFormationServiceMock = mock_services.NewMockCloudFormationServiceInterface(mockController)
			enableEBSCSIDriverInput = &EnableEBSCSIDriverInput{
				EKSService: eksServiceMock,
				IAMService: iamServiceMock,
				CFService:  cloudFormationServiceMock,
				Config:     &eksv1.EKSClusterConfig{},
			}
		})

		AfterEach(func() {
			mockController.Finish()
		})
		It("should detect that addon is already installed", func() {
			eksDescribeAddonOutput = &eks.DescribeAddonOutput{
				Addon: &eks.Addon{
					AddonArn: aws.String("arn:aws::ebs-csi-driver"),
				},
			}
			eksServiceMock.EXPECT().DescribeAddon(gomock.Any()).Return(eksDescribeAddonOutput, nil)
			addonArn, err := CheckEBSAddon(enableEBSCSIDriverInput.EKSService, enableEBSCSIDriverInput.Config)
			Expect(err).To(Succeed())
			Expect(addonArn).To(Equal("arn:aws::ebs-csi-driver"))
		})

		It("should detect that addon is not installed", func() {
			eksDescribeAddonOutput = &eks.DescribeAddonOutput{}
			eksServiceMock.EXPECT().DescribeAddon(gomock.Any()).Return(eksDescribeAddonOutput, nil)
			addonArn, err := CheckEBSAddon(enableEBSCSIDriverInput.EKSService, enableEBSCSIDriverInput.Config)
			Expect(err).To(Succeed())
			Expect(addonArn).To(Equal(""))
		})

		It("should fail to check if addon is not installed", func() {
			eksDescribeAddonOutput = &eks.DescribeAddonOutput{}
			eksServiceMock.EXPECT().DescribeAddon(gomock.Any()).Return(nil, fmt.Errorf("failed to describe addon"))
			_, err := CheckEBSAddon(enableEBSCSIDriverInput.EKSService, enableEBSCSIDriverInput.Config)
			Expect(err).ToNot(Succeed())
		})
	})
})
