package iam

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	"github.com/spectrocloud-labs/validator-plugin-aws/api/v1alpha1"
	"github.com/spectrocloud-labs/validator-plugin-aws/internal/utils/test"
	vapi "github.com/spectrocloud-labs/validator/api/v1alpha1"
	"github.com/spectrocloud-labs/validator/pkg/types"
	"github.com/spectrocloud-labs/validator/pkg/util/ptr"
)

type iamApiMock struct {
	attachedGroupPolicies map[string]*iam.ListAttachedGroupPoliciesOutput
	attachedRolePolicies  map[string]*iam.ListAttachedRolePoliciesOutput
	attachedUserPolicies  map[string]*iam.ListAttachedUserPoliciesOutput
	policyArns            map[string]*iam.GetPolicyOutput
	policyVersions        map[string]*iam.GetPolicyVersionOutput
}

func (m iamApiMock) GetPolicy(ctx context.Context, params *iam.GetPolicyInput, optFns ...func(*iam.Options)) (*iam.GetPolicyOutput, error) {
	return m.policyArns[*params.PolicyArn], nil
}

func (m iamApiMock) GetPolicyVersion(ctx context.Context, params *iam.GetPolicyVersionInput, optFns ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error) {
	return m.policyVersions[*params.PolicyArn], nil
}

func (m iamApiMock) ListAttachedGroupPolicies(ctx context.Context, params *iam.ListAttachedGroupPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedGroupPoliciesOutput, error) {
	return m.attachedGroupPolicies[*params.GroupName], nil
}

func (m iamApiMock) ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error) {
	return m.attachedRolePolicies[*params.RoleName], nil
}

func (m iamApiMock) ListAttachedUserPolicies(ctx context.Context, params *iam.ListAttachedUserPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedUserPoliciesOutput, error) {
	return m.attachedUserPolicies[*params.UserName], nil
}

const (
	policyDocumentOutput1 string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": [
					"ec2:DescribeInstances"
				],
				"Resource": [
					"*"
				],
				"Effect": "Allow"
			}
		]
	}`
	policyDocumentOutput2 string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": [
					"*"
				],
				"Resource": [
					"*"
				],
				"Effect": "Allow"
			}
		]
	}`
	policyDocumentOutput3 string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Condition": {
					"ForAnyValue:StringLike": {
						"kms:ResourceAliases": "alias/cluster-api-provider-aws-*"
					}
				},
				"Action": [
					"kms:CreateGrant",
					"kms:DescribeKey"
				],
				"Resource": [
					"*"
				],
				"Effect": "Allow"
			}
		]
	}`
	policyDocumentOutput4 string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"eks:AssociateIdentityProviderConfig",
					"eks:ListIdentityProviderConfigs"
				],
				"Resource": [
					"arn:*:eks:*:*:cluster/*"
				]
			},
			{
				"Effect": "Allow",
				"Action": [
					"eks:DisassociateIdentityProviderConfig",
					"eks:DescribeIdentityProviderConfig"
				],
				"Resource": [
					"*"
				]
			}
		]
	}`
	policyDocumentOutput5 string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": [
					"ec2:*",
					"s3:List*",
					"organizations:*Organizations",
					"iam:*Group*"
				],
				"Resource": [
					"*"
				],
				"Effect": "Allow"
			},
			{
				"Action": [
					"ec2:DescribeInstances"
				],
				"Resource": [
					"*"
				],
				"Effect": "Deny"
			}
		]
	}`
)

var iamService = NewIAMRuleService(logr.Logger{}, iamApiMock{
	attachedGroupPolicies: map[string]*iam.ListAttachedGroupPoliciesOutput{
		"iamGroup": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn1"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
	},
	attachedRolePolicies: map[string]*iam.ListAttachedRolePoliciesOutput{
		"iamRole1": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn1"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
		"iamRole2": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn2"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
		"iamRole3": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn3"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
		"iamRole4": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn4"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
		"iamRole5": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn5"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
	},
	policyArns: map[string]*iam.GetPolicyOutput{
		"iamRoleArn1": {
			Policy: ptr.Ptr(iamtypes.Policy{
				DefaultVersionId: ptr.Ptr("1"),
			}),
		},
		"iamRoleArn2": {
			Policy: ptr.Ptr(iamtypes.Policy{
				DefaultVersionId: ptr.Ptr("1"),
			}),
		},
		"iamRoleArn3": {
			Policy: ptr.Ptr(iamtypes.Policy{
				DefaultVersionId: ptr.Ptr("1"),
			}),
		},
		"iamRoleArn4": {
			Policy: ptr.Ptr(iamtypes.Policy{
				DefaultVersionId: ptr.Ptr("1"),
			}),
		},
		"iamRoleArn5": {
			Policy: ptr.Ptr(iamtypes.Policy{
				DefaultVersionId: ptr.Ptr("1"),
			}),
		},
	},
	attachedUserPolicies: map[string]*iam.ListAttachedUserPoliciesOutput{
		"iamUser": {
			AttachedPolicies: []iamtypes.AttachedPolicy{
				{
					PolicyArn:  ptr.Ptr("iamRoleArn1"),
					PolicyName: ptr.Ptr("iamPolicy"),
				},
			},
		},
	},
	policyVersions: map[string]*iam.GetPolicyVersionOutput{
		"iamRoleArn1": {
			PolicyVersion: ptr.Ptr(iamtypes.PolicyVersion{
				Document: ptr.Ptr(url.QueryEscape(policyDocumentOutput1)),
			}),
		},
		"iamRoleArn2": {
			PolicyVersion: ptr.Ptr(iamtypes.PolicyVersion{
				Document: ptr.Ptr(url.QueryEscape(policyDocumentOutput2)),
			}),
		},
		"iamRoleArn3": {
			PolicyVersion: ptr.Ptr(iamtypes.PolicyVersion{
				Document: ptr.Ptr(url.QueryEscape(policyDocumentOutput3)),
			}),
		},
		"iamRoleArn4": {
			PolicyVersion: ptr.Ptr(iamtypes.PolicyVersion{
				Document: ptr.Ptr(url.QueryEscape(policyDocumentOutput4)),
			}),
		},
		"iamRoleArn5": {
			PolicyVersion: ptr.Ptr(iamtypes.PolicyVersion{
				Document: ptr.Ptr(url.QueryEscape(policyDocumentOutput5)),
			}),
		},
	},
})

type testCase struct {
	name           string
	rule           iamRule
	expectedResult types.ValidationResult
	expectedError  error
}

func TestIAMGroupValidation(t *testing.T) {
	cs := []testCase{
		{
			name: "Fail (missing permission)",
			rule: v1alpha1.IamGroupRule{
				IamGroupName: "iamGroup",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"s3:GetBuckets"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-group-policy",
					ValidationRule: "validation-iamGroup",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamGroupRule iamGroup missing action(s): [s3:GetBuckets] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Pass (basic)",
			rule: v1alpha1.IamGroupRule{
				IamGroupName: "iamGroup",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"ec2:DescribeInstances"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-group-policy",
					ValidationRule: "validation-iamGroup",
					Message:        "All required aws-iam-group-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
	}
	for _, c := range cs {
		result, err := iamService.ReconcileIAMGroupRule(c.rule)
		test.CheckTestCase(t, result, c.expectedResult, err, c.expectedError)
	}
}

func TestIAMRoleValidation(t *testing.T) {
	cs := []testCase{
		{
			name: "Fail (missing permission)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole1",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"s3:GetBuckets"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole1",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamRoleRule iamRole1 missing action(s): [s3:GetBuckets] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Pass (basic)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole1",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"ec2:DescribeInstances"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole1",
					Message:        "All required aws-iam-role-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Pass (wildcard)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole2",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"ec2:DescribeInstances"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole2",
					Message:        "All required aws-iam-role-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Pass (condition)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole3",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Condition: &v1alpha1.Condition{
									Type:   "ForAnyValue:StringLike",
									Key:    "kms:ResourceAliases",
									Values: []string{"alias/cluster-api-provider-aws-*"},
								},
								Effect: "Allow",
								Actions: []string{
									"kms:CreateGrant",
									"kms:DescribeKey",
								},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole3",
					Message:        "All required aws-iam-role-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Fail (condition, missing value)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole3",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Condition: &v1alpha1.Condition{
									Type: "ForAnyValue:StringLike",
									Key:  "kms:ResourceAliases",
									Values: []string{
										"alias/cluster-api-provider-aws-*",
										"alias/another-value",
									},
								},
								Effect: "Allow",
								Actions: []string{
									"kms:CreateGrant",
									"kms:DescribeKey",
								},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole3",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"Condition ForAnyValue:StringLike: kms:ResourceAliases=[alias/cluster-api-provider-aws-* alias/another-value] not applied to action(s) [kms:CreateGrant kms:DescribeKey] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Fail (condition, total miss)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRole2",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Condition: &v1alpha1.Condition{
									Type:   "ForAnyValue:StringLike",
									Key:    "kms:ResourceAliases",
									Values: []string{"alias/cluster-api-provider-aws-*"},
								},
								Effect: "Allow",
								Actions: []string{
									"kms:CreateGrant",
									"kms:DescribeKey",
								},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRole2",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"Condition ForAnyValue:StringLike: kms:ResourceAliases=[alias/cluster-api-provider-aws-*] not applied to action(s) [kms:CreateGrant kms:DescribeKey] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Fail (error)",
			rule: v1alpha1.IamRoleRule{
				IamRoleName: "iamRoleZanzibar",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"kms:CreateGrant",
								},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-role-policy",
					ValidationRule: "validation-iamRoleZanzibar",
					Message:        "All required aws-iam-role-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
			expectedError: errors.New("no policies found for IAM role iamRoleZanzibar"),
		},
	}
	for _, c := range cs {
		result, err := iamService.ReconcileIAMRoleRule(c.rule)
		test.CheckTestCase(t, result, c.expectedResult, err, c.expectedError)
	}
}

func TestIAMUserValidation(t *testing.T) {
	cs := []testCase{
		{
			name: "Fail (missing permission)",
			rule: v1alpha1.IamUserRule{
				IamUserName: "iamUser",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"s3:GetBuckets"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-user-policy",
					ValidationRule: "validation-iamUser",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamUserRule iamUser missing action(s): [s3:GetBuckets] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Pass (basic)",
			rule: v1alpha1.IamUserRule{
				IamUserName: "iamUser",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"ec2:DescribeInstances"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-user-policy",
					ValidationRule: "validation-iamUser",
					Message:        "All required aws-iam-user-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
	}
	for _, c := range cs {
		result, err := iamService.ReconcileIAMUserRule(c.rule)
		test.CheckTestCase(t, result, c.expectedResult, err, c.expectedError)
	}
}

func TestIAMPolicyValidation(t *testing.T) {
	cs := []testCase{
		{
			name: "Fail (missing permission)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn1",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"s3:GetBuckets"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn1",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamPolicyRule iamRoleArn1 missing action(s): [s3:GetBuckets] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Pass (basic)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn1",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect:    "Allow",
								Actions:   []string{"ec2:DescribeInstances"},
								Resources: []string{"*"},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn1",
					Message:        "All required aws-iam-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Fail (multi-resource w/ wildcard)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn4",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"eks:AssociateIdentityProviderConfig",
									"eks:ListIdentityProviderConfigs",
									"eks:DisassociateIdentityProviderConfig",
									"eks:DescribeIdentityProviderConfig",
								},
								Resources: []string{
									"arn:*:eks:*:*:cluster/*",
									"arn:*:eks:*:*:nodegroup/*/*/*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn4",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamPolicyRule iamRoleArn4 missing action(s): [eks:AssociateIdentityProviderConfig eks:ListIdentityProviderConfigs] for resource arn:*:eks:*:*:nodegroup/*/*/* from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Fail (explicit deny override)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn5",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"ec2:DescribeInstances",
								},
								Resources: []string{
									"*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn5",
					Message:        "One or more required IAM permissions was not found, or a condition was not met",
					Details:        []string{},
					Failures: []string{
						"v1alpha1.IamPolicyRule iamRoleArn5 missing action(s): [ec2:DescribeInstances] for resource * from policy iamPolicy",
					},
					Status: corev1.ConditionFalse,
				},
				State: ptr.Ptr(vapi.ValidationFailed),
			},
		},
		{
			name: "Pass (explicit allow with irrelevant explicit deny)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn5",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"ec2:StartInstances",
								},
								Resources: []string{
									"*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn5",
					Message:        "All required aws-iam-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Pass (action with wildcard suffix)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn5",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"s3:ListBuckets",
								},
								Resources: []string{
									"*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn5",
					Message:        "All required aws-iam-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Pass (action with wildcard prefix)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn5",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"organizations:ListOrganizations",
								},
								Resources: []string{
									"*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn5",
					Message:        "All required aws-iam-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
		{
			name: "Pass (action with wildcard prefix and suffix)",
			rule: v1alpha1.IamPolicyRule{
				IamPolicyARN: "iamRoleArn5",
				Policies: []v1alpha1.PolicyDocument{
					{
						Name:    "iamPolicy",
						Version: "1",
						Statements: []v1alpha1.StatementEntry{
							{
								Effect: "Allow",
								Actions: []string{
									"iam:DetachGroupPolicy",
								},
								Resources: []string{
									"*",
								},
							},
						},
					},
				},
			},
			expectedResult: types.ValidationResult{
				Condition: &vapi.ValidationCondition{
					ValidationType: "aws-iam-policy",
					ValidationRule: "validation-iamRoleArn5",
					Message:        "All required aws-iam-policy permissions were found",
					Details:        []string{},
					Failures:       nil,
					Status:         corev1.ConditionTrue,
				},
				State: ptr.Ptr(vapi.ValidationSucceeded),
			},
		},
	}
	for _, c := range cs {
		result, err := iamService.ReconcileIAMPolicyRule(c.rule)
		test.CheckTestCase(t, result, c.expectedResult, err, c.expectedError)
	}
}
