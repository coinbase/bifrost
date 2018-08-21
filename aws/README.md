# AWS

This folder contains all the AWS logic.

`aws.go` builds the clients used by handlers to search, and create in AWS. It uses the `step` AWS clients to make it easy to assume roles into all accounts.
