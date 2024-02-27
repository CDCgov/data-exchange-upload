const axios = require("axios");

const c = {};

c.login = async (username, password, url) => {
  const params = new URLSearchParams({
    username: username,
    password: password,
  });

  return axios.post(`${url}/oauth`, params).then((response) => {
    if (response.status == 200 && response.statusText === "OK") {
      return response.data;
    }
    console.error(
      `Client login failed to SAMS, error code is ${response.status}, error message is ${response.statusText}`
    );
    return null;
  });
};

module.exports = c;
