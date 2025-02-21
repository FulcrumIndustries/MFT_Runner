const { app, port } = require('./server.js');

app.listen(port, () => {
    console.log(`Server started on port ${port}`);
}); 