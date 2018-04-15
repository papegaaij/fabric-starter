/**
 *
 */

module.exports = function (require) {

  let log4js = require('log4js');
  let logger = log4js.getLogger('orchestrator');

  const ORG = process.env.ORG || null;
  const USERNAME = process.env.SERVICE_USER || 'service' /*config.user.username*/;

  if(ORG !== 'ns' && ORG !== 'veolia') {
    logger.info('enabled for nsd only');
    return;
  }

  logger.info('**************    ORCHESTRATOR     ******************');
  logger.info('Admin   \t: ' + USERNAME);
  logger.info('**************                     ******************');

  let invoke = require('../lib-fabric/invoke-transaction.js');
  let query = require('../lib-fabric/query.js');
  let peerListener = require('../lib-fabric/peer-listener.js');

  logger.info('registering for block events');
  peerListener.registerBlockEvent(function (block) {
    try {
      block.data.data.forEach(blockData => {

        logger.info(`got block no. ${block.header.number}`);

        if (blockData.payload.data.actions) {
          blockData.payload.data.actions.forEach(action => {
            let extension = action.payload.action.proposal_response_payload.extension;
            let event = extension.events;
            logger.trace('event', JSON.stringify(event));
            if(!event.event_name) {
              return;
            }
            logger.trace(`event ${event.event_name}`);

            invoke.invokeChaincode(["grpcs://peer0.bank.transport-chain.nl:7051"],
              'bank-transport', 'payment', 'pay', [], USERNAME, ORG)
            .then(transactionId => {
              logger.info('invokeChaincode success ' + transactionId);
            })
            .catch(e => {
              logger.error('invokeChaincode', e);
            });

          }); // thru action elements
        }
      }); // thru block data elements
    }
    catch (e) {
      logger.error('Caught while processing block event', e);
    }
  });

};