const getBranchName = () => {
  return (
    process.env.CI_COMMIT_REF_NAME ||
    process.env.CI_COMMIT_BRANCH ||
    process.env.BRANCH_NAME ||
    'main'
  );
};

const getTagFormat = (branchName) => {
  if (branchName === 'main') {
    return `gosdk-v\${version}`;
  }
  return `gosdk-v\${version}-${branchName}`;
};

const currentBranch = getBranchName();

module.exports = {
  extends: 'semantic-release-monorepo',
  branches: ['main', 'staging'],
  repositoryUrl: 'https://gitlab.com/dreamsai/cpg-2/HeliumPay-monorepo',
  tagFormat: getTagFormat(currentBranch),
  plugins: [
    [
      '@semantic-release/commit-analyzer',
      {
        preset: 'angular',
        releaseRules: [
          {
            type: 'chore',
            release: 'patch',
          },
          {
            type: 'enh',
            release: 'patch',
          },
          {
            type: 'refactor',
            release: 'patch',
          },
          {
            type: 'hotfix',
            release: 'patch',
          },
        ],
      },
    ],
    '@semantic-release/release-notes-generator',
    [
      '@semantic-release/gitlab',
      {
        successCommentCondition: false,
      },
    ],
    [
      '@semantic-release/exec',
      {
        publishCmd:
          'cd ../../.. && GO_SDK_VERSION=${nextRelease.version} bash ci/scripts/deploy-go-sdk.sh',
      },
    ],
  ],
};
