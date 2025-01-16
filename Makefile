hello:
	@echo "hello world"
pt:
	@echo "生成协议..."
	cd ./protos && protoc --go_out=.  --go-grpc_out=. *.proto

pt_rpc:
	@echo "生成RPC协议..."
	cd ./protos &&  protoc --go_out=. --go-grpc_out=. *.proto