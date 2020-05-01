# Technical Debt
Known technical debt


## Saving a list
- Triggers a rebuild of all the users lists


## Hugo is a large binary
- The actual server is under 30mb, however including hugo pushes it above 70mb.
- Today building on a base allows us to store an image layer with hugo, reducing the upload.
- Tomorrow, it might be better to pull out hugo and have it as a requirement on the server, or
 have some way of sending the job to the services that is just a hugo static site generator.
