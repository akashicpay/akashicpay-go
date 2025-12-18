import createReleaseConfig from '../../../release.config.mjs';

export default createReleaseConfig({
  packageName: 'gosdk',
  publishCmd:
    'cd ../../.. && GO_SDK_VERSION=${nextRelease.version} bash ci/scripts/deploy-go-sdk.sh',
  prereleaseBranches: ['staging'],
});
