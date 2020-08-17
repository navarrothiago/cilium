// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/internal/awsutil"
)

// Contains the parameters for DescribeSpotDatafeedSubscription.
type DescribeSpotDatafeedSubscriptionInput struct {
	_ struct{} `type:"structure"`

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	DryRun *bool `locationName:"dryRun" type:"boolean"`
}

// String returns the string representation
func (s DescribeSpotDatafeedSubscriptionInput) String() string {
	return awsutil.Prettify(s)
}

// Contains the output of DescribeSpotDatafeedSubscription.
type DescribeSpotDatafeedSubscriptionOutput struct {
	_ struct{} `type:"structure"`

	// The Spot Instance data feed subscription.
	SpotDatafeedSubscription *SpotDatafeedSubscription `locationName:"spotDatafeedSubscription" type:"structure"`
}

// String returns the string representation
func (s DescribeSpotDatafeedSubscriptionOutput) String() string {
	return awsutil.Prettify(s)
}

const opDescribeSpotDatafeedSubscription = "DescribeSpotDatafeedSubscription"

// DescribeSpotDatafeedSubscriptionRequest returns a request value for making API operation for
// Amazon Elastic Compute Cloud.
//
// Describes the data feed for Spot Instances. For more information, see Spot
// Instance Data Feed (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-data-feeds.html)
// in the Amazon EC2 User Guide for Linux Instances.
//
//    // Example sending a request using DescribeSpotDatafeedSubscriptionRequest.
//    req := client.DescribeSpotDatafeedSubscriptionRequest(params)
//    resp, err := req.Send(context.TODO())
//    if err == nil {
//        fmt.Println(resp)
//    }
//
// Please also see https://docs.aws.amazon.com/goto/WebAPI/ec2-2016-11-15/DescribeSpotDatafeedSubscription
func (c *Client) DescribeSpotDatafeedSubscriptionRequest(input *DescribeSpotDatafeedSubscriptionInput) DescribeSpotDatafeedSubscriptionRequest {
	op := &aws.Operation{
		Name:       opDescribeSpotDatafeedSubscription,
		HTTPMethod: "POST",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &DescribeSpotDatafeedSubscriptionInput{}
	}

	req := c.newRequest(op, input, &DescribeSpotDatafeedSubscriptionOutput{})

	return DescribeSpotDatafeedSubscriptionRequest{Request: req, Input: input, Copy: c.DescribeSpotDatafeedSubscriptionRequest}
}

// DescribeSpotDatafeedSubscriptionRequest is the request type for the
// DescribeSpotDatafeedSubscription API operation.
type DescribeSpotDatafeedSubscriptionRequest struct {
	*aws.Request
	Input *DescribeSpotDatafeedSubscriptionInput
	Copy  func(*DescribeSpotDatafeedSubscriptionInput) DescribeSpotDatafeedSubscriptionRequest
}

// Send marshals and sends the DescribeSpotDatafeedSubscription API request.
func (r DescribeSpotDatafeedSubscriptionRequest) Send(ctx context.Context) (*DescribeSpotDatafeedSubscriptionResponse, error) {
	r.Request.SetContext(ctx)
	err := r.Request.Send()
	if err != nil {
		return nil, err
	}

	resp := &DescribeSpotDatafeedSubscriptionResponse{
		DescribeSpotDatafeedSubscriptionOutput: r.Request.Data.(*DescribeSpotDatafeedSubscriptionOutput),
		response:                               &aws.Response{Request: r.Request},
	}

	return resp, nil
}

// DescribeSpotDatafeedSubscriptionResponse is the response type for the
// DescribeSpotDatafeedSubscription API operation.
type DescribeSpotDatafeedSubscriptionResponse struct {
	*DescribeSpotDatafeedSubscriptionOutput

	response *aws.Response
}

// SDKResponseMetdata returns the response metadata for the
// DescribeSpotDatafeedSubscription request.
func (r *DescribeSpotDatafeedSubscriptionResponse) SDKResponseMetdata() *aws.Response {
	return r.response
}