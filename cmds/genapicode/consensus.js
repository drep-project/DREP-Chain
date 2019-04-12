
var Method = require('../method');
var formatters = require('../formatters');
var utils = require('../../utils/utils');

var CONSENSUS = function (drep) {
    this._requestManager = drep._requestManager;

    var self = this;
    
    methods().forEach(function(method) { 
        method.attachToObject(self);
        method.setRequestManager(drep._requestManager);
    });
};

var methods = function () {
	
var minning = new Method({
	name: 'minning',
	call: 'consensus_minning',
	params: 0,
});
	
var mock = new Method({
	name: 'mock',
	call: 'consensus_mock',
	params: 0,
});
	
    return [minning,mock]
}

module.exports = CONSENSUS;
