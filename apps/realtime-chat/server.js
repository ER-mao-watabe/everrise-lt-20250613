const express = require('express');
const http = require('http');
const socketIO = require('socket.io');
const path = require('path');

const app = express();
const server = http.createServer(app);
const io = socketIO(server);

const PORT = process.env.PORT || 3000;

// メッセージ履歴を保存する配列（簡易的な実装）
const messageHistory = [];
const MAX_HISTORY = 100;

// 静的ファイルの提供
app.use(express.static(path.join(__dirname, 'public')));

// Socket.IO接続処理
io.on('connection', (socket) => {
    console.log('新しいユーザーが接続しました');

    // 接続時に過去のメッセージ履歴を送信
    socket.emit('message history', messageHistory);

    // ユーザー名設定
    socket.on('set username', (username) => {
        socket.username = username;
        console.log(`ユーザー名設定: ${username}`);
        
        // 入室通知
        const joinMessage = {
            type: 'system',
            message: `${username}さんが入室しました`,
            timestamp: new Date()
        };
        io.emit('system message', joinMessage);
    });

    // メッセージ受信
    socket.on('chat message', (msg) => {
        if (!socket.username) {
            socket.emit('error', 'ユーザー名を設定してください');
            return;
        }

        const messageData = {
            username: socket.username,
            message: msg,
            timestamp: new Date()
        };

        // メッセージ履歴に追加
        messageHistory.push(messageData);
        if (messageHistory.length > MAX_HISTORY) {
            messageHistory.shift();
        }

        // 全クライアントにメッセージを送信
        io.emit('chat message', messageData);
    });

    // 切断処理
    socket.on('disconnect', () => {
        if (socket.username) {
            const leaveMessage = {
                type: 'system',
                message: `${socket.username}さんが退室しました`,
                timestamp: new Date()
            };
            io.emit('system message', leaveMessage);
        }
        console.log('ユーザーが切断しました');
    });
});

server.listen(PORT, () => {
    console.log(`サーバーがポート ${PORT} で起動しました`);
});