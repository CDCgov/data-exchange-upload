import dotenv from "dotenv";
dotenv.config({ path: "../../.env" });

interface EnvConfig {
  validateEnv: () => [string, string, string, string | undefined];
}

const config: EnvConfig = {
  validateEnv: () => {
    // Use CI/CD environment variables or fallback to .env variables
    const username = process.env.SAMS_USERNAME || process.env.ACCOUNT_USERNAME;
    const password = process.env.SAMS_PASSWORD || process.env.ACCOUNT_PASSWORD;
    const url = process.env.UPLOAD_URL || process.env.DEX_URL;
    const ps_url = process.env.PS_API_URL;

    let error = "No ";
    let hasErrors = false;

    if (username == null || username === "") {
      hasErrors = true;
      error += "username";
    }

    if (password == null || password === "") {
      if (hasErrors) {
        error += ", password";
      } else {
        error += "password";
      }
      hasErrors = true;
    }

    if (url == null || url === "") {
      if (hasErrors) {
        error += ", data exchange url";
      } else {
        error += "data exchange url";
      }
      hasErrors = true;
    }

    if (hasErrors) {
      error += " has been set in environment.";
      console.error(`Terminating environment: ${error}`);
      process.exit(1);
    }

    return [username!, password!, url!, ps_url];
  },
};

export default config;

