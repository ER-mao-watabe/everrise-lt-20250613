const express = require('express');
const cors = require('cors');
const fs = require('fs');
const path = require('path');

const app = express();
const PORT = 3000;

app.use(cors());
app.use(express.json());
app.use(express.static('public'));

// 売上データを読み込む
const salesData = JSON.parse(
  fs.readFileSync(path.join(__dirname, 'data', 'sales-data.json'), 'utf8')
);

// API: 全売上データを取得
app.get('/api/sales', (req, res) => {
  res.json(salesData);
});

// API: 月別売上データを取得
app.get('/api/sales/monthly', (req, res) => {
  const monthly = salesData.salesData.map(item => ({
    month: item.month,
    revenue: item.revenue
  }));
  res.json(monthly);
});

// API: 商品別売上ランキングを取得
app.get('/api/sales/product-ranking', (req, res) => {
  const productTotals = {};
  
  salesData.salesData.forEach(month => {
    month.products.forEach(product => {
      if (!productTotals[product.name]) {
        productTotals[product.name] = { quantity: 0, amount: 0 };
      }
      productTotals[product.name].quantity += product.quantity;
      productTotals[product.name].amount += product.amount;
    });
  });
  
  const ranking = Object.entries(productTotals)
    .map(([name, data]) => ({ name, ...data }))
    .sort((a, b) => b.amount - a.amount);
    
  res.json(ranking);
});

// API: 売上予測を取得（簡単な線形回帰）
app.get('/api/sales/forecast', (req, res) => {
  const monthly = salesData.salesData.map((item, index) => ({
    x: index + 1,
    y: item.revenue
  }));
  
  // 線形回帰の計算
  const n = monthly.length;
  const sumX = monthly.reduce((sum, item) => sum + item.x, 0);
  const sumY = monthly.reduce((sum, item) => sum + item.y, 0);
  const sumXY = monthly.reduce((sum, item) => sum + item.x * item.y, 0);
  const sumX2 = monthly.reduce((sum, item) => sum + item.x * item.x, 0);
  
  const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;
  
  // 次の3ヶ月の予測
  const forecast = [];
  for (let i = 1; i <= 3; i++) {
    const x = n + i;
    const revenue = Math.round(slope * x + intercept);
    const month = new Date(2024, 5 + i, 1).toISOString().slice(0, 7);
    forecast.push({ month, revenue, predicted: true });
  }
  
  res.json({
    historical: salesData.salesData,
    forecast,
    model: { slope, intercept }
  });
});

// API: フィルタリングされた売上データを取得
app.get('/api/sales/filtered', (req, res) => {
  const { startMonth, endMonth, product } = req.query;
  let filteredData = [...salesData.salesData];
  
  if (startMonth) {
    filteredData = filteredData.filter(item => item.month >= startMonth);
  }
  
  if (endMonth) {
    filteredData = filteredData.filter(item => item.month <= endMonth);
  }
  
  if (product) {
    filteredData = filteredData.map(month => ({
      ...month,
      products: month.products.filter(p => p.name === product),
      revenue: month.products
        .filter(p => p.name === product)
        .reduce((sum, p) => sum + p.amount, 0)
    }));
  }
  
  res.json(filteredData);
});

app.listen(PORT, () => {
  console.log(`売上ダッシュボードサーバーが http://localhost:${PORT} で起動しました`);
});