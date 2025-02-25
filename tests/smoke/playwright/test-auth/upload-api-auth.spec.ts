import { expect, test, request, BrowserContext, Cookie} from '@playwright/test';
import { CookieOptions } from 'express';


test.describe('Upload API Auth UI Elements', () => {
    const pages = [
        {
            pageName: "root UI page",
            path: "/",
            expectedRedirect: ""
        },
        {
            pageName: "login page",
            path: "/login",
            expectedRedirect: ""
        },
        {
            pageName: "manifest page",
            path: "/manifest?data_stream_id=dextesting&data_stream_route=testevent1",
            expectedRedirect: "?redirect=/manifest?data_stream_id%3Ddextesting%26data_stream_route%3Dtestevent1"
        }
    ]
    pages.forEach(({ pageName, path, expectedRedirect }) => {
        test(`requires a token for ${pageName}`, async ({ page, baseURL }) => {
            await page.goto(path)

            await expect(page.getByRole('heading', {name: "Welcome to PHDO Upload Login"})).toBeVisible()
            await expect(page.getByRole('textbox', {name: "Authentication Token *"})).toBeVisible()
            await expect(page.getByRole('button', {name: "Login"})).toBeVisible()
            
            await expect(page.url()).toBe(`${baseURL}/login${expectedRedirect}`)
            
        })

    })
})

test.describe('Upload API Auth', () => {
    let token: any
    let expiredToken: any

    test.beforeAll(async ({ request }) => {
        const res = await request.get('http://localhost:3000/token')
        const json = await res.json()
        token = json.access_token


        const resExpired = await request.get('http://localhost:3000/token-expired')
        const jsonExpired = await resExpired.json()
        expiredToken = jsonExpired.access_token
    })
    
    test('logs in with a valid token through the login page', async ({ page }) => {
        await page.goto('/')  
        await page.getByRole('textbox', { name: "Authentication Token *" }).fill(token)
        await page.getByRole('button', { name: 'Login' }).click()
        await expect(page.getByRole('heading', { name: 'Welcome to DEX Upload' })).toBeVisible()
        await expect(page.getByText("Start the upload process by entering a data stream and route.")).toBeVisible()
    })

    test('cannot log in with an invalid token through the login page', async ({ page }) => {
        await page.goto('/')  
        await page.getByRole('textbox', { name: "Authentication Token *" }).fill("badtoken")
        await page.getByRole('button', { name: 'Login' }).click()
        await expect(page.url()).toContain("/login")
        await expect(page.getByText("Welcome to PHDO Upload Login")).toBeVisible()
    })

    test('cannot log in with an expired token through the login page', async ({ page }) => { 
        await page.goto('/')  
        await page.getByRole('textbox', { name: "Authentication Token *" }).fill(expiredToken)
        await page.getByRole('button', { name: 'Login' }).click()
        await expect(page.url()).toContain("/login")
        await expect(page.getByText("Welcome to PHDO Upload Login")).toBeVisible()
    })
    
    test('logs in with a valid cookie token', async ({ browser }) => {
        const cookies = [{
            name: "phdo_auth_token",
            value: token,
            domain: "localhost",
            path: "/",
            expires: Math.floor(Date.now() / 1000) + 3600,
            httpOnly: true,
            secure: true,
            sameSite: "Lax" as const
        }]
        const context = await browser.newContext();
        context.addCookies(cookies)
        const page = await context.newPage();
        await page.goto('/');
        await expect(page.getByRole('heading', { name: 'Welcome to DEX Upload' })).toBeVisible()
        await expect(page.getByText("Start the upload process by entering a data stream and route.")).toBeVisible()
    })

    test('cannot log in with an expired cookie token', async ({browser}) => {
        const cookies = [{
            name: "phdo_auth_token",
            value: token,
            domain: "localhost",
            path: "/",
            expires: Math.floor(Date.now() / 1000) - 1,
            httpOnly: true,
            secure: true,
            sameSite: "Lax" as const
        }]
        const context = await browser.newContext();
        context.addCookies(cookies)
        const page = await context.newPage();
        await page.goto('/');
        await expect(page.url()).toContain("/login")
        await expect(page.getByText("Welcome to PHDO Upload Login")).toBeVisible()

    })

    test('cannot log in with an invalid cookie token', async ({browser}) => {
        const cookies = [{
            name: "phdo_auth_token",
            value: "badtoken",
            domain: "localhost",
            path: "/",
            expires: Math.floor(Date.now() / 1000) + 3600,
            httpOnly: true,
            secure: true,
            sameSite: "Lax" as const
        }]
        const context = await browser.newContext();
        context.addCookies(cookies)
        const page = await context.newPage();
        await page.goto('/');
        await expect(page.url()).toContain("/login")
        await expect(page.getByText("Welcome to PHDO Upload Login")).toBeVisible()

    })
    
})

