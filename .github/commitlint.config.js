module.exports = {
  extends: ["@commitlint/config-conventional"],
  helpUrl:
    "https://github.com/kumahq/kuma/blob/master/CONTRIBUTING.md#commit-message-format",
  rules: {
    "body-max-line-length": [0],
    "footer-max-line-length": [0],
    "footer-leading-blank": [0],
    "header-max-length": [0],
  },
};
