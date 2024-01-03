#!/bin/sh
 #遍历所有的proto源文件
for file in cases/*.proto
do
    arr=(${arr[*]} $file)
   #每一个proto文件执行一次
   protoc --go_out=paths=source_relative:. \
   --go-validate_out=paths=source_relative:. \
   --proto_path=${HOME}/github.com/ml444/gkit/cmd/protoc-gen-go-validate \
   --proto_path=${HOME}/github.com/ml444/gkit/cmd/protoc-gen-go-validate/tests \
   --proto_path=${HOME}/github.com/bryce4651/gctl/cmd/protos \
   $file
   # protoc —go-validate_out=gocases/ $file
done
#输出遍历结果
echo  ${arr[@]}

go fmt ./cases/*.pb.go
goimports -w ./cases/*.pb.go

# for pbfile in ./gocases/*_validate.pb.go
# do
#     pbArr=(${pbArr[*]} $pbfile)
# done
# #输出转换结果 
# echo ${pbArr[@]}

#将转换后的文件移动到新文件夹
# mkdir pb_file
# mv -f -v *_validate.pb.go pb_file

