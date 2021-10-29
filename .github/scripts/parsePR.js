module.exports = async ({
  github = {},
  owner = 'kumahq',
  repo = 'kuma',
}, foundPR = null) => {
  if (!foundPR) {
    return null
  }

  try {
    const { number, title } = foundPR;

    const { data = {} } = await github.rest.pulls.get({
      owner,
      repo,
      pull_number: number,
    });

    const {
      user = {},
      merged_by: mergedBy = {},
      labels: fullLabels = [],
    } = data;

    const labels = fullLabels.map(({ id, name, description }) => ({
      id,
      name,
      description,
    }));

    return {
      number: number,
      title: title,
      openedBy: {
        login: user.login,
        id: user.id,
      },
      mergedBy: {
        login: mergedBy.login,
        id: mergedBy.id,
      },
      labels,
    };
  } catch (e) {
    console.error(e);
  }

  return null;
};
