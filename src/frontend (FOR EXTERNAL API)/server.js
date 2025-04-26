const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const path = require('path');

const app = express();
const PORT = 3000;

app.use(express.static(path.join(__dirname, 'public')));

app.use('/api', createProxyMiddleware({
  target: 'https://arlan-api.azurewebsites.net',
  changeOrigin: true,
  pathRewrite: {
    '^/api': '', 
  },
}));

app.listen(PORT, () => {
  console.log(`Сервер запущен на http://localhost:${PORT}`);
});
