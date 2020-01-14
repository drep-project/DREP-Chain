1 使用genaccount产生一份配置文件

2 修改config.json中的producerNum的值为:drep公链部署时所有出块节点的个数。此值可向drep官方联系获取。

3 替换config.json中的"genesis"字段的值为:drep公链部署时，所用的"genesis"的值，此值可向drep官方联系获取。

4 修改config.json中的字段 "StaticNodes": [	 ], 添加任意一个公链上的enode字符，例如  "enode://20104fe34bd96f38d6b74964066139d109307e163914516b8cf283de0def0e1b@192.168.31.220:44444"； 或者使用本地api接口p2p_addPeer来完成。

5 获取出块账号，此账号就是使用genaccount产生的keystore下的一个文件名称；pubkey就是config.json中字段mypk对应的pubkey; 使用api:p2p_localNode获取出块账号所在节点的p2p信息，同时保证出块账号有足够的余额。（注意,如果不想使用genaccount生成的原始账户,可以使用其他账户，要保证config.json的中mypk和质押币时的pubkey相同）

6 调用account_candidateCredit，把上述获取到的信息账户地址\地址对应的pubkey\本地p2p enode信息，填写到具体位置即可。
