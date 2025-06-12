const socket = io();

let username = '';

// DOM要素
const usernameModal = document.getElementById('usernameModal');
const usernameInput = document.getElementById('usernameInput');
const setUsernameBtn = document.getElementById('setUsernameBtn');
const chatContainer = document.getElementById('chatContainer');
const usernameDisplay = document.getElementById('usernameDisplay');
const messagesContainer = document.getElementById('messages');
const messageForm = document.getElementById('messageForm');
const messageInput = document.getElementById('messageInput');

// ユーザー名設定
setUsernameBtn.addEventListener('click', () => {
    const inputUsername = usernameInput.value.trim();
    if (inputUsername && inputUsername.length > 0) {
        username = inputUsername;
        socket.emit('set username', username);
        usernameModal.style.display = 'none';
        chatContainer.style.display = 'flex';
        usernameDisplay.textContent = `👤 ${username}`;
        messageInput.focus();
    }
});

// Enterキーでユーザー名設定
usernameInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        setUsernameBtn.click();
    }
});

// メッセージ送信
messageForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const message = messageInput.value.trim();
    if (message) {
        socket.emit('chat message', message);
        messageInput.value = '';
        messageInput.focus();
    }
});

// メッセージ受信
socket.on('chat message', (data) => {
    displayMessage(data);
});

// システムメッセージ受信
socket.on('system message', (data) => {
    displaySystemMessage(data.message);
});

// メッセージ履歴受信
socket.on('message history', (history) => {
    messagesContainer.innerHTML = '';
    history.forEach(message => {
        displayMessage(message);
    });
});

// エラーハンドリング
socket.on('error', (error) => {
    alert(error);
});

// メッセージ表示
function displayMessage(data) {
    const messageEl = document.createElement('div');
    messageEl.className = 'message';
    
    const time = new Date(data.timestamp).toLocaleTimeString('ja-JP', {
        hour: '2-digit',
        minute: '2-digit'
    });
    
    messageEl.innerHTML = `
        <div class="message-header">
            <span class="message-username">${escapeHtml(data.username)}</span>
            <span class="message-time">${time}</span>
        </div>
        <div class="message-content">${escapeHtml(data.message)}</div>
    `;
    
    messagesContainer.appendChild(messageEl);
    scrollToBottom();
}

// システムメッセージ表示
function displaySystemMessage(message) {
    const messageEl = document.createElement('div');
    messageEl.className = 'system-message';
    messageEl.textContent = message;
    messagesContainer.appendChild(messageEl);
    scrollToBottom();
}

// 絵文字挿入
function insertEmoji(emoji) {
    messageInput.value += emoji;
    messageInput.focus();
}

// 最下部にスクロール
function scrollToBottom() {
    messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

// HTMLエスケープ
function escapeHtml(text) {
    const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, m => map[m]);
}

// ページロード時にユーザー名入力欄にフォーカス
window.addEventListener('load', () => {
    usernameInput.focus();
});