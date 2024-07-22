# TODO's

[ ] Add github action to build code and run tests
[x] Serve public dir to provide css, js and images to the frontend
[x] Create Handler type for serving views and a seperate one for serving data
[x] Create DB schema for user data storage
[ ] Consider storing images outside the repo to reduce project size
[x] Add users email to site cookie, and store it in req context when a message happens
[x] Style login screen
[x] Style Chat windows
[x] Store messages in a database so they can be persited when a user leaves or joins a chat
[x] Firgure out how to handle multiiple chat rooms
[x] Allow users to be invited to a chat
[x] Install HTMX locally (store HTMX file in /js, or install with NPM/Node)
[ ] Migrate to Postgres + Reddis Database
[x] IsAuthed middleware to ensure that a user is authed before allowing them on certain routes
[ ] Create logger for use within project to assist with consistency
[ ] Use SQLC to generate and manage database queries. GORM is fine for small things, but becomes messy with more complex queries. Note: bun could also be investigated in conjunction with goose.
[ ] Upgrade to HTMX v2 <!-- *This should be done soon as it is core the the applications function -->

[x] Feat: Login
[x] When a user logs in certain fields should / should not be shown (login button, etc)
[x] Add auth with login redirect to routes that need it
[x] Add link to register from login page and reverse

[ ] Feat: Chat
[x] Dont render messages with no/invisible content
[ ] Add emoji support
[ ] Scroll bar should move to the bottom of the chat window when a HTMX request is completed
[x] Ensure rooms display all data correctly
[ ] flesh out the invitation logic a bit more, needs to return something when an invite is sent

[ ] Feat: Team space
[ ] Create create team page
[ ] Create page to display a users teams
[ ] Auth for team only chats
[ ] Each team should have a chat associated with it

[ ] Feat: User profile
[ ] Create page
[ ] User should have a profile image
[ ] User should be able to modify the details of their profile from this page

[ ] Feat: Error handling
[ ] Create an error page that can be displayed when neccisary
[ ] Create some error cards/divs that can be used in various scenarios to indicate where failures have occured

[ ] Feat
[ ] Create Config package
[x] Create .env file to store sensitive data

[ ] Feat: Live video streaming with WebRTC (might need a webcam to make this work)
[ ] Create a page to display a users camera feed and a chat
