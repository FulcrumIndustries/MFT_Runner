const express = require('express');
const path = require('path');

const app = express();
const port = process.env.PORT || 3000;

// Serve static files from dist directory
app.use(express.static(path.join(__dirname, 'dist')));

// Handle SPA fallback
app.get('*', (req, res) => {
    res.sendFile(path.join(__dirname, 'dist', 'index.html'));
});

app.listen(port, () => {
    console.log(`MFT Runner UI serving from: ${path.join(__dirname, 'dist')}`);
    console.log(`Server running on port ${port}`);
    console.log(`You can use the following URLs to access the application:`);
    console.log(`- http://localhost:${port}`);
    console.log(`- http://127.0.0.1:${port}`);
}); 