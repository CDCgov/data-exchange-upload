import c, { LoginResponse } from './client';
import config from './config';

export async function loginAndGetToken() {
    const [username, password, url] = config.validateEnv();

    const loginResponse: LoginResponse | null = await c.login(username, password, url);

    if (loginResponse === null) {
        throw new Error("Login failed. Exiting program.");
    }

    return loginResponse.access_token;
}
