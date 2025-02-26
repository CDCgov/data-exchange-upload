require('dotenv').config();
const express = require('express');
const cors = require('cors');
const crypto = require('crypto');
const jwt = require('jsonwebtoken');
const base64url = require('base64url');

const app = express();
const PORT = process.env.MOCK_AUTH_PORT || 3000;
const BASE_URL = `${process.env.MOCK_AUTH_BASE_URL}:${PORT}` || `http://localhost:${PORT}`;

app.use(cors());
app.use(express.json());

let jwks = { keys: [] };
let privateKeyPem = "";

const validPayload = {
    sub: "1234567890",
    name: "JOHN DOE",
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + 3600, // 1-hour expiration
    iss: `${BASE_URL}`,
    aud: [
        `${BASE_URL}/*`
    ],
    jti: crypto.randomUUID(),
    email: "testemail@nowhere",
    account_type: "system",
    account_id: "SYS-99999",
    family_name: "DOE",
    middle_name: "X",
    given_name: "JOHN",
    preferred_name: "JOHN",
    scope: "test:scope1 test:scope2",
    jku: `${BASE_URL}/openid/connect/jwks.json`
};

// Function to generate RSA key pair and JWKS
function generateKeys() {
    const { publicKey, privateKey } = crypto.generateKeyPairSync("rsa", {
        modulusLength: 2048,
        publicKeyEncoding: { type: "spki", format: "pem" },
        privateKeyEncoding: { type: "pkcs8", format: "pem" }
    });

    privateKeyPem = privateKey; // Store private key for signing JWTs

    // Extract key details
    const publicKeyObj = crypto.createPublicKey(publicKey);
    const keyDetails = publicKeyObj.export({ format: 'jwk' });

    jwks = {
        keys: [
            {
                kty: keyDetails.kty,
                kid: crypto.randomUUID(),
                n: base64url.encode(Buffer.from(keyDetails.n, 'base64')),
                e: base64url.encode(Buffer.from(keyDetails.e, 'base64'))
            }
        ]
    };
    console.log("Private Key: ", privateKey)
    console.log("Public Key: ", publicKey)
    console.log("Generated JWKS:", jwks);
}

generateKeys(); // Generate RSA keys at startup

// OpenID Configuration Endpoint
app.get('/.well-known/openid-configuration', (req, res) => {
    res.json({
        issuer: `${BASE_URL}`,
        authorization_endpoint: `${BASE_URL}/oauth2/authorize`,
        token_endpoint: `${BASE_URL}/oauth2/token`,
        jwks_uri: `${BASE_URL}/oauth2/jwks`,
    });
});

// JWKS Endpoint
app.get('/oauth2/jwks', (req, res) => {
    res.json(jwks);
});

// Token Endpoint (Issues JWT)
app.get('/token', (req, res) => {
    // Find the key ID from JWKS
    const keyId = jwks.keys[0].kid;

    // Sign JWT using the private RSA key
    const token = jwt.sign(validPayload, privateKeyPem, {
        algorithm: "RS256",
        keyid: keyId
    });

    res.json({ access_token: token, token_type: "Bearer", expires_in: 3600 });
});

app.get('/token-expired', (req, res) => {
    const invalidPayload = { ...validPayload }
    invalidPayload.exp = Math.floor(Date.now() / 1000) - 1

    // Find the key ID from JWKS
    const keyId = jwks.keys[0].kid;

    // Sign JWT using the private RSA key
    const token = jwt.sign(invalidPayload, privateKeyPem, {
        algorithm: "RS256",
        keyid: keyId
    });

    res.json({ access_token: token, token_type: "Bearer", expires_in: 3600 });
});


app.get('/token-scopes', (req, res) => {
    const invalidPayload = { ...validPayload }
    invalidPayload.scope = "test:scope3 test:scope4"

    // Find the key ID from JWKS
    const keyId = jwks.keys[0].kid;

    // Sign JWT using the private RSA key
    const token = jwt.sign(invalidPayload, privateKeyPem, {
        algorithm: "RS256",
        keyid: keyId
    });

    res.json({ access_token: token, token_type: "Bearer", expires_in: 3600 });
});

app.listen(PORT, () => {
    console.log(`Mock OpenID server running at ${BASE_URL}`);
});
