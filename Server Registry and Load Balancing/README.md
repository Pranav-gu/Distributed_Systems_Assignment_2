# Distributed Systems Assignment 2
## Name - Pranav Gupta
## Roll No. - 2021101095



### Assumptions:
The directory structure of the question is as follows:
1. lb_server folder: Contains code for load balancing server.
2. client folder: Contains code fpr normal execution of client(not for scale-testing).
3. client_test folder: Contains code catering to scale-testing. Not required to run it locally on systems, and only required for scle-testing.
4. server folder: contains the source code for all the backend servers.
5. protofiles: Contains the services.proto file which contains the declaring of RPCs.
6. analyze.py: Python script to draw graphs and perform analysis.
7. Report.pdf listing all the Steps followed, assumptions stated and the executed procedure alongwith the graphs for this question.


### Execution:
Follow these steps to run the code.
go mod init Q1
go mod tidy
go run ./lb_server/main.go <policy>. Policy can be “least_loadˮ, “pick_firstˮ,
“round_robinˮ.
go run ./server/main.go —port=<port no.>, each server on different terminal
go run ./client/main.go

In order to perform scale-testing, execute the following command:

./scale_script1.sh --servers=15 --clients=100 --requests=20.

This automaticallly spwans multiple processes and generate graphs and detailed analysis in a folder.