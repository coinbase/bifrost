# EC2

This folder contains the example EC2 client for AWS.

An AWS clients goal is to abstract the specific AWS calls into a easier to understand API. The goal is so that the main code depends on this AWS package and not directly AWS's.

For example, we redefine `ec2.Instance` here to add easer methods like `Helathy()` so that the release and handler code is easier to understand.
