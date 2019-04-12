
var Method = require('../method');
var formatters = require('../formatters');
var utils = require('../../utils/utils');

var LOG = function (drep) {
    this._requestManager = drep._requestManager;

    var self = this;
    
    methods().forEach(function(method) { 
        method.attachToObject(self);
        method.setRequestManager(drep._requestManager);
    });
};

var methods = function () {
	
var setBackTraceAt = new Method({
	name: 'setBackTraceAt',
	call: 'log_setBackTraceAt',
	params: 1,
});
	
var setLevel = new Method({
	name: 'setLevel',
	call: 'log_setLevel',
	params: 1,
});
	
var setVmodule = new Method({
	name: 'setVmodule',
	call: 'log_setVmodule',
	params: 1,
});
	
    return [setBackTraceAt,setLevel,setVmodule]
}

module.exports = LOG;
