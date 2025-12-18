const createReleaseConfig = require('../../../release.config.js');

module.exports = createReleaseConfig({
  packageName: 'gosdk',
  publishCmd:
    'cd ../../.. && GO_SDK_VERSION=${nextRelease.version} bash ci/scripts/deploy-go-sdk.sh',
  prereleaseBranches: ['staging'],
});
