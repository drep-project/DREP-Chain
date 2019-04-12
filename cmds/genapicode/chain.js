
var Method = require('../method');
var formatters = require('../formatters');
var utils = require('../../utils/utils');

var CHAIN = function (drep) {
    this._requestManager = drep._requestManager;

    var self = this;
    
    methods().forEach(function(method) { 
        method.attachToObject(self);
        method.setRequestManager(drep._requestManager);
    });
};

var methods = function () {
	
var getBalance = new Method({
	name: 'getBalance',
	call: 'chain_getBalance',
	params: 1,
});
	
var getBlock = new Method({
	name: 'getBlock',
	call: 'chain_getBlock',
	params: 1,
});
	
var getMaxHeight = new Method({
	name: 'getMaxHeight',
	call: 'chain_getMaxHeight',
	params: 0,
	outputFormatter : utils.toDecimal
});
	
var getNonce = new Method({
	name: 'getNonce',
	call: 'chain_getNonce',
	params: 1,
	outputFormatter : utils.toDecimal
});
	
var getPreviousBlockHash = new Method({
	name: 'getPreviousBlockHash',
	call: 'chain_getPreviousBlockHash',
	params: 0,
});
	
var getReputation = new Method({
	name: 'getReputation',
	call: 'chain_getReputation',
	params: 1,
});
	
var getTransactionByBlockHeightAndIndex = new Method({
	name: 'getTransactionByBlockHeightAndIndex',
	call: 'chain_getTransactionByBlockHeightAndIndex',
	params: 2,
});
	
var getTransactionCountByBlockHeight = new Method({
	name: 'getTransactionCountByBlockHeight',
	call: 'chain_getTransactionCountByBlockHeight',
	params: 1,
});
	
var getTransactionsFromBlock = new Method({
	name: 'getTransactionsFromBlock',
	call: 'chain_getTransactionsFromBlock',
	params: 1,
});
	
var sendRawTransaction = new Method({
	name: 'sendRawTransaction',
	call: 'chain_sendRawTransaction',
	params: 1,
});
	
    return [getBalance,getBlock,getMaxHeight,getNonce,getPreviousBlockHash,getReputation,getTransactionByBlockHeightAndIndex,getTransactionCountByBlockHeight,getTransactionsFromBlock,sendRawTransaction]
}

module.exports = CHAIN;
