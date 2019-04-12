
var Method = require('../method');
var formatters = require('../formatters');
var utils = require('../../utils/utils');

var P2P = function (drep) {
    this._requestManager = drep._requestManager;

    var self = this;
    
    methods().forEach(function(method) { 
        method.attachToObject(self);
        method.setRequestManager(drep._requestManager);
    });
};

var methods = function () {
	
var addPeers = new Method({
	name: 'addPeers',
	call: 'p2p_addPeers',
	params: 1,
});
	
var getPeers = new Method({
	name: 'getPeers',
	call: 'p2p_getPeers',
	params: 0,
});
	
var removePeers = new Method({
	name: 'removePeers',
	call: 'p2p_removePeers',
	params: 1,
});
	
    return [addPeers,getPeers,removePeers]
}

module.exports = P2P;
