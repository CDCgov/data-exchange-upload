import axios from "axios";

export interface LoginResponse {
  // Define the structure of your login response here
  // Example:
  access_token: string;
  // Add other fields as necessary
}

const c = {
  async login(username: string, password: string, url: string): Promise<LoginResponse | null> {
    const params = new URLSearchParams({
      username: username,
      password: password,
    });

    try {
            
      const response = await axios.post(`${url}/oauth`, params);

      console.log("Raw response data:", response.data); 

      if (response.status === 200 && response.statusText === "OK") {
        return response.data as LoginResponse;
      } else {
        console.error(
          `Client login failed to SAMS, error code is ${response.status}, error message is ${response.statusText}`
        );
        return null;
      }
    } catch (error) {
      console.error("An error occurred during login:", error);
      return null;
    }
  },
};

export default c;

