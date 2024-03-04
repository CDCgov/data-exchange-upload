export interface LoginResponse {
  access_token: string;
}

const c = {
  async login(username: string, password: string, url: string): Promise<LoginResponse | null> {
    const params = new URLSearchParams({
      username: username,
      password: password,
    });

    try {
      const response = await fetch(`${url}/oauth`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: params,
      });

      if (!response.ok) {
        console.error(
          `Client login failed to SAMS, error code is ${response.status}, error message is ${response.statusText}`
        );
        return null;
      }

      const data: LoginResponse = await response.json();
      return data;
    } catch (error) {
      console.error("An error occurred during login:", error);
      return null;
    }
  },
};

export default c;

