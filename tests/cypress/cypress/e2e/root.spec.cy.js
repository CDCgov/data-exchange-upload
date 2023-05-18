describe("root", () => {
  it("should return 200", () => {
    cy.api("GET", Cypress.env("DEX_URL")).then(resp => {
      expect(resp.status).to.eq(200)
    })
  })
})
