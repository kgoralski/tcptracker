### Level 1:

> 1. How would you prove the code is correct?

_Answer_: Unit testing, manual testing, testing in isolation on CI (maybe also integration testing) and in some testing environment. Having good enough test coverage and tests with race conditions.
The directory `tools/pings.sh` also got small script to run to check port scanning detection. Added Github Actions CI.

> 2. How would you make this solution better?

_Answer_: It's better to use the `pcap` library instead of reading the file every 10 seconds, like on the Level 1. I have added the Github Actions. Publish monitoring data to monitoring systems like Prometheus add some alerts if needed. Publish the docker container to image registry. I'm using the GoCache in my solution so using different backends under the hood would be also possible, or even chaining them. For real life scenario we would need probably more sophisticated Firewall Rules. I'm not firewall expert but looking at the different library during this task I would take look for `nftables` or `firewalld` that I think can support multiple backends. I could also find a better building tool than Makefile. Another thing would be the releasing the artifact process and publishing it.

> 3. Is it possible for this program to miss a connection?

_Answer_: Yes, same like the `tcpdump` tool it can suffer "When tcpdump "drops" packets, is because from it has not enough buffer space to keep up with the packets arriving from the network"

> 4. If you weren't following these requirements, how would you solve the problem of logging every new connection?

_Answer_: I would think about the business requirements. In this scenario we are mostly logging the new connection. Maybe adding this data to the monitoring system with some label would be more meaningful. We can build the automation around the logs, but usually it's not something that a lot of tooling exists for. 
With adding it to prometheus we could integrate that data with a lot of solutions. Some concern would be if Prometheus would keep up with new traffic and scale and growing size. 
I'm quite sure there are existing solutions out there to use. Maybe worth checking fail2ban, wireguard, tailscale, but that's more about banning functionality. 

### Level 2:

> 1. Why did you choose x to write the build automation?

_Answer_: Mostly simplicity and availability in nearly every OS. For this task it was good enough because it helps during the development process with having fast feedback loops. For longer term for sure it would not be enough. 

> 2. Is there anything else you would test if you had more time?

_Answer_: Yes, I would test more for concurrency, race conditions. I would add some benchmarks. I would also check the application with different profilers, https://etcnotes.com/posts/pprof/ to check for unnecessary allocations, memory etc. It should be built into Jetbrains Goland right now. 

> 3. What is the most important tool, script, or technique you have for solving problems in production? Explain why this tool/script/technique is the most important.

_Answer_: IMO being able to reproduce the production issue quickly on different e.g. Staging environment is one of the most important techniques. Quite often solving a bug can be about 90% of time trying to reproduce it and 10% it's when we can finally easily track it. Sometimes Engineers that can properly debug something cannot have access to production environment because of security policies, PII data, PCI scope etc., which makes it even more important. Usually it is also quite hard to create such environment.
This is also why I find unit testing and integration testing quite important, basic thing that gives fast feedback loops. Without them, we typically need to test everything in production which makes it much slower.

### Level 3:

> 1. If you had to deploy this program to hundreds of servers, what would be your preferred method? Why?

_Answer_: I would add CI/CD automation with proper building and releasing artifact flow and pushing Docker image to docker registry.
My preference would be to deploy it using some Cloud Container services - my preference would be because it's the most familiar service for me.

> 2. What is the hardest technical problem or outage you've had to solve in your career? Explain what made it so difficult?

_Answer_: Maybe not the hardest in my career but something that happened in last year. so I can remember it better.
Randomness. The issue was related to the AWS Lambda that was running inside the VPC, at some day lambda started to throw 500 errors, redeployed and started to work but not for long, after few hours it was "fixed". Going back to work next day it happened again. 
I have started to investigate anything I could and with AWS Lambda let's say it's quite "limited". Checked everything in the network, enabled AWS X-Ray, yes it's indeed just lambda, which sometimes can just start to throw the errors, compared different setups, everything started to look fine. 
Also checked that other lambdas, and it looked that only in that one team is a problem. Until I have finally spotted that also others can have similar issues. 
Finally,  I have realized that maybe somebody was deleting AWS resources recently. 

This way I was able to find in AWS Cloudtrail a massive detach of some AWS Lambda Roles `AWSLambdaBasicExecutionRole`, `AWSLambdaVPCAccessExecutionRole`, because it was a "clean up" few days ago. 
The weird thing was the redeployment of the stack never helped, the Serverless Framework has never failed in the deployment process. 
We have added the roles to the stacks, and it started to work properly again.

Description of the role:
> AWSLambdaVPCAccessExecutionRole – Grants permissions for Amazon Elastic Compute Cloud (Amazon EC2) actions to manage elastic network interfaces (ENIs). If you are writing a Lambda function to access resources in a VPC in the Amazon Virtual Private Cloud (Amazon VPC) service, you can attach this permissions policy. The policy also grants permissions for CloudWatch Logs actions to write logs.

Also, when I have joined current company it was freshly migrated to AWS and everything was on fire, and we had few  war rooms, RabbitMQ, Mysql, Monolith etc. A lot of configuration drifts, Connection reliability improvements, keep alives, timeouts were implemented. 

With other examples I would say that when I was working on bare metal it was quite hard and probably harder, for example: as a small team we were maintaining PostgreSQL replication, or having Apache Solr inside the Docker with Zookepers and "Hello, you got two leaders inside the cluster" was also weird. 
For a few years I work with the Cloud, and it's a blessing than running everything on our own. 

### Level 4:

-- no questions in this section