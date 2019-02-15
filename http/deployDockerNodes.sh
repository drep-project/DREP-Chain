#!/bin/bash 
# ~/WorkSpace/scripts/deployDockerNodes.sh main 4 backup 3 55550
# ~/WorkSpace/scripts/deployDockerNodes.sh drepcli 7 backup 0 55551
# ~/WorkSpace/scripts/deployDockerNodes.sh drepcli 4 logsyncfix 0 55551

name=$1
num=$2
branch=$3
ip=$4
port=$5

if [[ ! $name || ! $num || ! $branch || ! $ip || ! $port ]]; then
	i=0
	if [ ! $name ]; then
		echo "need name param!"
		let i++
	elif [ ! $num ]; then
		echo "need num param!"
		let i++
	elif [ ! $branch ]; then
		echo "need branch param!"
		let i++
	elif [ ! $ip ]; then
		echo "need ip param!"
		let i++
	elif [ ! $port ]; then
		echo "need port param!"
		let i++
	fi
	echo "shell at least need 5 params but got $i !"
else
	cd ~/go/src
	
	# git clone "https://github.com/drep-project/bc.git"
	# username="hjw6160602@163.com"
	# spawn git pull
	# expect "Username for 'https://github.com':"
	# send "$username\n"
	# interact
	# mv bc BlockChainTest

	cd BlockChainTest
	git checkout $branch

	if [[ $name == "drepcli" ]]; then
		cd  cli/${name}
	fi
	
	go install
	cd ~/go/bin
	if [[ $name == "main" ]]; then
		mv BlockChainTest main
	fi

	cp $name ~/DREP/${name}.${num}p/node.eric/
	echo "copy $name to node.eric succeed!"
	cp $name ~/DREP/${name}.${num}p/node.xie/
	echo "copy $name to node.xie succeed!"
	cp $name ~/DREP/${name}.${num}p/node.sai/
	echo "copy $name to node.sai succeed!"
	cp $name ~/DREP/${name}.${num}p/node.zbu/
	echo "copy $name to node.zbu succeed!"
	if [[ $num == 7 ]]; then
		cp $name ~/DREP/${name}.${num}p/node.long/
		echo "copy $name to node.long succeed!"
	
		cp $name ~/DREP/${name}.${num}p/node.hei/
		echo "copy $name to node.hei succeed!"
	
		cp $name ~/DREP/${name}.${num}p/node.backup/
		echo "copy $name to node.backup succeed!"
	fi

	echo "docker is building ${num} node version, tag: $branch.${num}p ..."

	cd ~/DREP/${name}.${num}p/node.eric;
	docker image build -t node.eric:$branch.${num}p .;
	
	cd ~/DREP/${name}.${num}p/node.xie;
	docker image build -t node.xie:$branch.${num}p .;
	
	cd ~/DREP/${name}.${num}p/node.sai;
	docker image build -t node.sai:$branch.${num}p .;
	
	cd ~/DREP/${name}.${num}p/node.zbu;
	docker image build -t node.zbu:$branch.${num}p .;
	
	if [[ $num == 7 ]]; then
		cd ~/DREP/${name}.${num}p/node.hei;
		docker image build -t node.hei:$branch.${num}p .;
	
		cd ~/DREP/${name}.${num}p/node.long;
		docker image build -t node.long:$branch.${num}p .;
	
		cd ~/DREP/${name}.${num}p/node.backup;
		docker image build -t node.backup:$branch.${num}p .;
	fi

	network="Wifi${ip}F"
	echo "docker is run tag: $branch ..."

	docker run --network=${network} --ip 192.168.${ip}.108 -d node.xie:$branch.${num}p;
    docker run --network=${network} --ip 192.168.${ip}.141 -d node.sai:$branch.${num}p;
    docker run --network=${network} --ip 192.168.${ip}.111 -d node.zbu:$branch.${num}p;
    if [[ $num == 7 ]]; then
		docker run --network=${network} --ip 192.168.${ip}.142 -d node.hei:$branch.${num}p;
    	docker run --network=${network} --ip 192.168.${ip}.198 -d node.long:$branch.${num}p;
    	docker run --network=${network} --ip 192.168.${ip}.119 -d node.backup:$branch.${num}p;
	fi
	docker run --network=${network} --ip 192.168.${ip}.152 -d -p ${port}:55550 node.eric:$branch.${num}p;
fi

