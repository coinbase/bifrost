package deployer

import (
	"github.com/coinbase/bifrost/aws"
	"github.com/coinbase/step/handler"
	"github.com/coinbase/step/machine"
)

// StateMachine returns the StateMachine for the deployer
func StateMachine() (*machine.StateMachine, error) {
	return machine.FromJSON([]byte(`{
    "Comment": "Bifrost Deployer",
    "StartAt": "Validate",
    "States": {
      "Validate": {
        "Type": "TaskFn",
        "Comment": "Validate and Set Defaults",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Lock",
        "Catch": [
          {
            "Comment": "Bad Release or Error GoTo end",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          }
        ]
      },
      "Lock": {
        "Type": "TaskFn",
        "Comment": "Grab Lock",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "ValidateResources",
        "Catch": [
          {
            "Comment": "Something else is deploying",
            "ErrorEquals": ["LockExistsError"],
            "ResultPath": "$.error",
            "Next": "FailureClean"
          },
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "CleanUpFailure"
          }
        ]
      },
      "ValidateResources": {
        "Type": "TaskFn",
        "Comment": "ValidateResources",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Deploy",
        "Catch": [
          {
            "Comment": "Try Release Lock Then Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "CleanUpFailure"
          }
        ]
      },
      "Deploy": {
        "Type": "TaskFn",
        "Comment": "Upload Step-Function and Lambda",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "WaitForDeploy",
        "Catch": [
          {
            "Comment": "Unsure of State, Fail",
            "ErrorEquals": ["States.ALL"],
            "ResultPath": "$.error",
            "Next": "CleanUpFailure"
          }
        ]
      },
      "WaitForDeploy": {
        "Comment": "Give the Deploy time to boot instance",
        "Type": "Wait",
        "Seconds" : 30,
        "Next": "WaitForHealthy"
      },
      "WaitForHealthy": {
        "Type": "Wait",
        "Seconds" : 15,
        "Next": "CheckHealthy"
      },
      "CheckHealthy": {
        "Type": "TaskFn",
        "Comment": "Is the new deploy healthy? Should we continue checking?",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Healthy?",
        "Retry": [{
          "Comment": "Do not retry on HaltError",
          "ErrorEquals": ["HaltError"],
          "MaxAttempts": 0
        },
        {
          "Comment": "Retry a few times in case of another error",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 15
        }],
        "Catch": [{
          "Comment": "Fail",
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "CleanUpFailure"
        }]
      },
      "Healthy?": {
        "Comment": "Check the release is $.healthy",
        "Type": "Choice",
        "Choices": [
          {
            "Comment": "Continue to Success",
            "Variable": "$.healthy",
            "BooleanEquals": true,
            "Next": "CleanUpSuccess"
          },
          {
            "Comment": "Loop Back and check if it is healthy again",
            "Variable": "$.healthy",
            "BooleanEquals": false,
            "Next": "WaitForHealthy"
          }
        ],
        "Default": "CleanUpFailure"
      },
      "CleanUpFailure": {
        "Type": "TaskFn",
        "Comment": "Release the Lock and Fail",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "FailureClean",
        "Retry": [ {
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "CleanUpSuccess": {
        "Type": "TaskFn",
        "Comment": "Release the Lock and Destroy Old instance",
        "Resource": "arn:aws:lambda:{{aws_region}}:{{aws_account}}:function:{{lambda_name}}",
        "Next": "Success",
        "Retry": [ {
          "Comment": "Keep trying to Release",
          "ErrorEquals": ["States.ALL"],
          "MaxAttempts": 3,
          "IntervalSeconds": 30
        }],
        "Catch": [{
          "ErrorEquals": ["States.ALL"],
          "ResultPath": "$.error",
          "Next": "FailureDirty"
        }]
      },
      "FailureClean": {
        "Comment": "Deploy Failed Cleanly",
        "Type": "Fail",
        "Error": "NotifyError"
      },
      "FailureDirty": {
        "Comment": "Deploy Failed, Resources left in Bad State, ALERT!",
        "Type": "Fail",
        "Error": "AlertError"
      },
      "Success": {
        "Type": "Succeed"
      }
    }
  }`))
}

// TaskHandlers returns
func TaskHandlers() *handler.TaskHandlers {
	return CreateTaskHandlers(&aws.ClientsStr{})
}

// CreateTaskHandlers returns
func CreateTaskHandlers(awsc aws.Clients) *handler.TaskHandlers {
	tm := handler.TaskHandlers{}
	tm["Validate"] = Validate(awsc)
	tm["Lock"] = Lock(awsc)
	tm["ValidateResources"] = ValidateResources(awsc)
	tm["Deploy"] = Deploy(awsc)
	tm["CleanUpFailure"] = CleanUpFailure(awsc)
	tm["CleanUpSuccess"] = CleanUpSuccess(awsc)
	tm["CheckHealthy"] = CheckHealthy(awsc)

	return &tm
}
