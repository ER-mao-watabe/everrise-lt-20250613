// グローバル変数
let monthlyChart, rankingChart, forecastChart;
let currentData = null;

// ページ読み込み時の初期化
document.addEventListener('DOMContentLoaded', async () => {
    await loadAllData();
    setupEventListeners();
});

// イベントリスナーの設定
function setupEventListeners() {
    document.getElementById('applyFilter').addEventListener('click', applyFilters);
    document.getElementById('resetFilter').addEventListener('click', resetFilters);
}

// 全データの読み込み
async function loadAllData() {
    try {
        // 月別売上データ
        const monthlyResponse = await fetch('/api/sales/monthly');
        const monthlyData = await monthlyResponse.json();
        createMonthlyChart(monthlyData);

        // 商品別ランキング
        const rankingResponse = await fetch('/api/sales/product-ranking');
        const rankingData = await rankingResponse.json();
        createRankingChart(rankingData);

        // 売上予測
        const forecastResponse = await fetch('/api/sales/forecast');
        const forecastData = await forecastResponse.json();
        createForecastChart(forecastData);

        // サマリー情報の更新
        updateSummary(monthlyData);
    } catch (error) {
        console.error('データの読み込みエラー:', error);
    }
}

// 月別売上グラフの作成
function createMonthlyChart(data) {
    const ctx = document.getElementById('monthlyChart').getContext('2d');
    
    if (monthlyChart) {
        monthlyChart.destroy();
    }

    monthlyChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: data.map(item => formatMonth(item.month)),
            datasets: [{
                label: '月別売上高',
                data: data.map(item => item.revenue),
                borderColor: '#3498db',
                backgroundColor: 'rgba(52, 152, 219, 0.1)',
                borderWidth: 3,
                tension: 0.4,
                pointRadius: 5,
                pointHoverRadius: 7,
                pointBackgroundColor: '#3498db',
                pointBorderColor: '#fff',
                pointBorderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            return `売上高: ¥${context.parsed.y.toLocaleString()}`;
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '¥' + value.toLocaleString();
                        }
                    }
                }
            },
            interaction: {
                mode: 'index',
                intersect: false
            }
        }
    });
}

// 商品別ランキングチャートの作成
function createRankingChart(data) {
    const ctx = document.getElementById('rankingChart').getContext('2d');
    
    if (rankingChart) {
        rankingChart.destroy();
    }

    rankingChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: data.map(item => item.name),
            datasets: [{
                label: '売上金額',
                data: data.map(item => item.amount),
                backgroundColor: [
                    'rgba(231, 76, 60, 0.8)',
                    'rgba(46, 204, 113, 0.8)',
                    'rgba(52, 152, 219, 0.8)',
                    'rgba(155, 89, 182, 0.8)',
                    'rgba(241, 196, 15, 0.8)'
                ],
                borderColor: [
                    'rgba(231, 76, 60, 1)',
                    'rgba(46, 204, 113, 1)',
                    'rgba(52, 152, 219, 1)',
                    'rgba(155, 89, 182, 1)',
                    'rgba(241, 196, 15, 1)'
                ],
                borderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const item = data[context.dataIndex];
                            return [
                                `売上金額: ¥${item.amount.toLocaleString()}`,
                                `販売数量: ${item.quantity.toLocaleString()}個`
                            ];
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '¥' + value.toLocaleString();
                        }
                    }
                }
            }
        }
    });
}

// 売上予測チャートの作成
function createForecastChart(data) {
    const ctx = document.getElementById('forecastChart').getContext('2d');
    
    if (forecastChart) {
        forecastChart.destroy();
    }

    // 実績と予測データを結合
    const allData = [...data.historical, ...data.forecast];
    const labels = allData.map(item => formatMonth(item.month));
    
    // 実績データ
    const historicalData = data.historical.map(item => item.revenue);
    // 予測データ（実績の最後の点から開始）
    const forecastData = new Array(data.historical.length - 1).fill(null);
    forecastData.push(data.historical[data.historical.length - 1].revenue);
    data.forecast.forEach(item => forecastData.push(item.revenue));

    forecastChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '実績',
                data: historicalData,
                borderColor: '#3498db',
                backgroundColor: 'rgba(52, 152, 219, 0.1)',
                borderWidth: 3,
                tension: 0.4,
                pointRadius: 5
            }, {
                label: '予測',
                data: forecastData,
                borderColor: '#e74c3c',
                backgroundColor: 'rgba(231, 76, 60, 0.1)',
                borderWidth: 3,
                borderDash: [5, 5],
                tension: 0.4,
                pointRadius: 5
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    position: 'top'
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.dataset.label;
                            return `${label}: ¥${context.parsed.y.toLocaleString()}`;
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        callback: function(value) {
                            return '¥' + value.toLocaleString();
                        }
                    }
                }
            }
        }
    });

    // 予測情報の表示
    const forecastInfo = document.getElementById('forecastInfo');
    forecastInfo.innerHTML = `
        <p><strong>予測モデル情報:</strong></p>
        <p>線形回帰による売上予測（傾き: ${data.model.slope.toFixed(2)}, 切片: ${data.model.intercept.toFixed(2)}）</p>
        <p>※ 過去6ヶ月のデータを基に、今後3ヶ月の売上を予測しています。</p>
    `;
}

// サマリー情報の更新
function updateSummary(monthlyData) {
    // 合計売上
    const totalRevenue = monthlyData.reduce((sum, item) => sum + item.revenue, 0);
    document.getElementById('totalRevenue').textContent = '¥' + totalRevenue.toLocaleString();

    // 月平均売上
    const avgRevenue = totalRevenue / monthlyData.length;
    document.getElementById('avgRevenue').textContent = '¥' + Math.round(avgRevenue).toLocaleString();

    // 最高売上月
    const maxMonth = monthlyData.reduce((max, item) => 
        item.revenue > max.revenue ? item : max
    );
    document.getElementById('maxMonth').textContent = 
        formatMonth(maxMonth.month) + ' (¥' + maxMonth.revenue.toLocaleString() + ')';

    // 成長率（最初の月と最後の月の比較）
    if (monthlyData.length >= 2) {
        const firstMonth = monthlyData[0].revenue;
        const lastMonth = monthlyData[monthlyData.length - 1].revenue;
        const growthRate = ((lastMonth - firstMonth) / firstMonth * 100).toFixed(1);
        document.getElementById('growthRate').textContent = growthRate + '%';
    }
}

// フィルター適用
async function applyFilters() {
    const startMonth = document.getElementById('startMonth').value;
    const endMonth = document.getElementById('endMonth').value;
    const product = document.getElementById('productFilter').value;

    const params = new URLSearchParams();
    if (startMonth) params.append('startMonth', startMonth);
    if (endMonth) params.append('endMonth', endMonth);
    if (product) params.append('product', product);

    try {
        const response = await fetch('/api/sales/filtered?' + params);
        const filteredData = await response.json();
        
        // 月別データの更新
        const monthlyData = filteredData.map(item => ({
            month: item.month,
            revenue: item.revenue
        }));
        createMonthlyChart(monthlyData);
        updateSummary(monthlyData);

        // 商品フィルターが適用されている場合は、ランキングチャートも更新
        if (product) {
            const productData = [];
            filteredData.forEach(month => {
                month.products.forEach(p => {
                    const existing = productData.find(item => item.name === p.name);
                    if (existing) {
                        existing.amount += p.amount;
                        existing.quantity += p.quantity;
                    } else {
                        productData.push({ ...p });
                    }
                });
            });
            createRankingChart(productData);
        }
    } catch (error) {
        console.error('フィルター適用エラー:', error);
    }
}

// フィルターリセット
function resetFilters() {
    document.getElementById('startMonth').value = '2024-01';
    document.getElementById('endMonth').value = '2024-06';
    document.getElementById('productFilter').value = '';
    loadAllData();
}

// 月表示のフォーマット
function formatMonth(month) {
    const date = new Date(month + '-01');
    return date.toLocaleDateString('ja-JP', { year: 'numeric', month: 'long' });
}