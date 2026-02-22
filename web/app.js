const BOARD_SIZE = 15;
const CELL_SIZE = 36;
const PADDING = 21;

let ws = null;
let currentUser = null;
let currentRoom = null;
let currentGame = null;
let myColor = 0;
let board = [];

const MessageType = {
    Ping: 1000,
    Pong: 1001,
    Login: 2001,
    LoginResp: 2002,
    Register: 2003,
    RegisterResp: 2004,
    CreateRoom: 3001,
    CreateRoomResp: 3011,
    JoinRoom: 3002,
    JoinRoomResp: 3012,
    LeaveRoom: 3003,
    LeaveRoomResp: 3013,
    RoomList: 3004,
    RoomListResp: 3014,
    RoomInfo: 3005,
    PlayerJoin: 3015,
    PlayerLeave: 3016,
    Move: 4001,
    MoveResp: 4002,
    GameOver: 4003,
    GameStart: 4004,
    BoardUpdate: 4005,
    ForfeitReq: 4006,
    ForfeitResp: 4007,
    LeaderboardReq: 5001,
    LeaderboardResp: 5002,
    UserStatsReq: 5003,
    UserStatsResp: 5004,
    Error: 9999
};

function connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${protocol}//${window.location.host}/ws`);

    ws.onopen = () => {
        console.log('WebSocket connected');
        const token = localStorage.getItem('token');
        if (token) {
            send(MessageType.Login, { token: token });
        }
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        handleMessage(data.type, data.payload);
    };

    ws.onclose = () => {
        console.log('WebSocket disconnected');
        setTimeout(connect, 3000);
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

function send(type, payload) {
    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type, payload }));
    }
}

function handleMessage(type, payload) {
    switch (type) {
        case MessageType.LoginResp:
            handleLoginResp(payload);
            break;
        case MessageType.RegisterResp:
            handleRegisterResp(payload);
            break;
        case MessageType.CreateRoomResp:
            handleCreateRoomResp(payload);
            break;
        case MessageType.JoinRoomResp:
            handleJoinRoomResp(payload);
            break;
        case MessageType.LeaveRoomResp:
            handleLeaveRoomResp(payload);
            break;
        case MessageType.RoomListResp:
            handleRoomListResp(payload);
            break;
        case MessageType.PlayerJoin:
            handlePlayerJoin(payload);
            break;
        case MessageType.PlayerLeave:
            handlePlayerLeave(payload);
            break;
        case MessageType.GameStart:
            handleGameStart(payload);
            break;
        case MessageType.BoardUpdate:
            handleBoardUpdate(payload);
            break;
        case MessageType.GameOver:
            handleGameOver(payload);
            break;
        case MessageType.LeaderboardResp:
            handleLeaderboardResp(payload);
            break;
        case MessageType.UserStatsResp:
            handleUserStatsResp(payload);
            break;
        case MessageType.Error:
            alert(payload.message);
            break;
    }
}

function showTab(tab) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.form').forEach(f => f.classList.add('hidden'));
    
    event.target.classList.add('active');
    document.getElementById(`${tab}-form`).classList.remove('hidden');
}

function showPage(pageId) {
    document.querySelectorAll('.page').forEach(p => p.classList.add('hidden'));
    document.getElementById(pageId).classList.remove('hidden');
}

function login() {
    const username = document.getElementById('login-username').value.trim();
    const password = document.getElementById('login-password').value;
    
    if (!username || !password) {
        alert('请输入用户名和密码');
        return;
    }
    
    send(MessageType.Login, { username, password });
}

function register() {
    const username = document.getElementById('reg-username').value.trim();
    const password = document.getElementById('reg-password').value;
    
    if (!username || !password) {
        alert('请输入用户名和密码');
        return;
    }
    
    send(MessageType.Register, { username, password });
}

function handleLoginResp(payload) {
    if (payload.code === 200) {
        localStorage.setItem('token', payload.token);
        currentUser = { id: payload.user_id, token: payload.token };
        showPage('lobby-page');
        document.getElementById('username-display').textContent = `用户: ${payload.user_id}`;
        send(MessageType.UserStatsReq, { user_id: payload.user_id });
        send(MessageType.RoomList, {});
        send(MessageType.LeaderboardReq, { limit: 10 });
    } else {
        alert(payload.message);
    }
}

function handleRegisterResp(payload) {
    if (payload.code === 200) {
        alert('注册成功，请登录');
        showTab('login');
    } else {
        alert(payload.message);
    }
}

function handleUserStatsResp(payload) {
    if (payload.code === 200) {
        document.getElementById('score-display').textContent = `积分: ${payload.score}`;
    }
}

function logout() {
    localStorage.removeItem('token');
    currentUser = null;
    showPage('auth-page');
}

function createRoom() {
    const roomName = prompt('请输入房间名称:', `${currentUser.id}的房间`);
    if (roomName) {
        send(MessageType.CreateRoom, { room_name: roomName });
    }
}

function handleCreateRoomResp(payload) {
    if (payload.code === 200) {
        currentRoom = { id: payload.room_id };
        showPage('room-page');
        document.getElementById('room-name').textContent = `房间 ${payload.room_id}`;
        document.getElementById('player-1').querySelector('.player-name').textContent = currentUser.id;
        document.getElementById('player-2').querySelector('.player-name').textContent = '等待中...';
    } else {
        alert(payload.message);
    }
}

function joinRoom(roomId) {
    send(MessageType.JoinRoom, { room_id: roomId });
}

function handleJoinRoomResp(payload) {
    if (payload.code === 200) {
        currentRoom = { id: payload.room_id };
        showPage('room-page');
        document.getElementById('room-name').textContent = `房间 ${payload.room_id}`;
        send(MessageType.RoomList, {});
    } else {
        alert(payload.message);
    }
}

function leaveRoom() {
    send(MessageType.LeaveRoom, { room_id: currentRoom.id });
}

function handleLeaveRoomResp(payload) {
    currentRoom = null;
    showPage('lobby-page');
    send(MessageType.RoomList, {});
}

function handleRoomListResp(payload) {
    const roomList = document.getElementById('room-list');
    roomList.innerHTML = '';
    
    if (payload.rooms && payload.rooms.length > 0) {
        payload.rooms.forEach(room => {
            const div = document.createElement('div');
            div.className = 'room-item';
            const isFull = room.players && room.players.length >= 2;
            div.innerHTML = `
                <div class="room-info">
                    <span class="room-name">${room.room_name}</span>
                    <span class="room-players">玩家: ${room.players ? room.players.length : 0}/2</span>
                </div>
                <button onclick="joinRoom(${room.room_id})" ${isFull ? 'disabled' : ''}>
                    ${isFull ? '已满' : '加入'}
                </button>
            `;
            roomList.appendChild(div);
        });
    } else {
        roomList.innerHTML = '<p style="color: rgba(255,255,255,0.5); text-align: center;">暂无房间</p>';
    }
}

function handlePlayerJoin(payload) {
    document.getElementById('player-2').querySelector('.player-name').textContent = payload.username;
    document.getElementById('room-status').textContent = '玩家已加入，等待游戏开始...';
}

function handlePlayerLeave(payload) {
    document.getElementById('player-2').querySelector('.player-name').textContent = '等待中...';
    document.getElementById('room-status').textContent = '对手已离开，等待新玩家...';
}

function handleLeaderboardResp(payload) {
    const leaderboard = document.getElementById('leaderboard');
    leaderboard.innerHTML = '';
    
    if (payload.ranks && payload.ranks.length > 0) {
        payload.ranks.forEach((rank, index) => {
            const div = document.createElement('div');
            div.className = 'rank-item';
            div.innerHTML = `
                <span class="rank-number">${index + 1}</span>
                <div class="rank-info">
                    <span class="rank-name">${rank.username}</span>
                    <span class="rank-score">${rank.score}分 | 胜率 ${rank.win_rate}</span>
                </div>
            `;
            leaderboard.appendChild(div);
        });
    }
}

function handleGameStart(payload) {
    currentGame = {
        roomId: payload.room_id,
        players: payload.players,
        currentPlayer: payload.first_player
    };
    
    const myIndex = payload.players.indexOf(currentUser.id);
    myColor = myIndex === 0 ? 1 : 2;
    
    board = Array(BOARD_SIZE).fill(null).map(() => Array(BOARD_SIZE).fill(0));
    
    showPage('game-page');
    initBoard();
    updateTurnInfo();
}

function initBoard() {
    const canvas = document.getElementById('game-board');
    const ctx = canvas.getContext('2d');
    
    ctx.fillStyle = '#dcb35c';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    
    ctx.strokeStyle = '#000';
    ctx.lineWidth = 1;
    
    for (let i = 0; i < BOARD_SIZE; i++) {
        ctx.beginPath();
        ctx.moveTo(PADDING + i * CELL_SIZE, PADDING);
        ctx.lineTo(PADDING + i * CELL_SIZE, PADDING + (BOARD_SIZE - 1) * CELL_SIZE);
        ctx.stroke();
        
        ctx.beginPath();
        ctx.moveTo(PADDING, PADDING + i * CELL_SIZE);
        ctx.lineTo(PADDING + (BOARD_SIZE - 1) * CELL_SIZE, PADDING + i * CELL_SIZE);
        ctx.stroke();
    }
    
    const starPoints = [[3, 3], [3, 11], [11, 3], [11, 11], [7, 7]];
    ctx.fillStyle = '#000';
    starPoints.forEach(([x, y]) => {
        ctx.beginPath();
        ctx.arc(PADDING + x * CELL_SIZE, PADDING + y * CELL_SIZE, 4, 0, 2 * Math.PI);
        ctx.fill();
    });
    
    canvas.onclick = handleCanvasClick;
}

function handleCanvasClick(event) {
    if (!currentGame || currentGame.currentPlayer !== currentUser.id) {
        return;
    }
    
    const canvas = document.getElementById('game-board');
    const rect = canvas.getBoundingClientRect();
    const x = Math.round((event.clientX - rect.left - PADDING) / CELL_SIZE);
    const y = Math.round((event.clientY - rect.top - PADDING) / CELL_SIZE);
    
    if (x >= 0 && x < BOARD_SIZE && y >= 0 && y < BOARD_SIZE && board[x][y] === 0) {
        send(MessageType.Move, { room_id: currentRoom.id, x, y });
    }
}

function handleBoardUpdate(payload) {
    board = payload.board;
    currentGame.currentPlayer = payload.current_player;
    
    drawBoard();
    updateTurnInfo();
}

function drawBoard() {
    const canvas = document.getElementById('game-board');
    const ctx = canvas.getContext('2d');
    
    initBoard();
    
    for (let i = 0; i < BOARD_SIZE; i++) {
        for (let j = 0; j < BOARD_SIZE; j++) {
            if (board[i][j] !== 0) {
                drawStone(ctx, i, j, board[i][j]);
            }
        }
    }
}

function drawStone(ctx, x, y, color) {
    const cx = PADDING + x * CELL_SIZE;
    const cy = PADDING + y * CELL_SIZE;
    const radius = CELL_SIZE / 2 - 2;
    
    ctx.beginPath();
    ctx.arc(cx, cy, radius, 0, 2 * Math.PI);
    
    if (color === 1) {
        const gradient = ctx.createRadialGradient(cx - 5, cy - 5, 0, cx, cy, radius);
        gradient.addColorStop(0, '#555');
        gradient.addColorStop(1, '#000');
        ctx.fillStyle = gradient;
    } else {
        const gradient = ctx.createRadialGradient(cx - 5, cy - 5, 0, cx, cy, radius);
        gradient.addColorStop(0, '#fff');
        gradient.addColorStop(1, '#ccc');
        ctx.fillStyle = gradient;
    }
    
    ctx.fill();
    ctx.strokeStyle = color === 1 ? '#000' : '#999';
    ctx.lineWidth = 1;
    ctx.stroke();
}

function updateTurnInfo() {
    const turnInfo = document.getElementById('turn-info');
    const isMyTurn = currentGame.currentPlayer === currentUser.id;
    
    if (isMyTurn) {
        turnInfo.textContent = '轮到你了！';
        turnInfo.style.color = '#4ecca3';
    } else {
        turnInfo.textContent = '等待对手落子...';
        turnInfo.style.color = '#e94560';
    }
    
    document.getElementById('current-turn').textContent = 
        `你是${myColor === 1 ? '黑' : '白'}方 | ${isMyTurn ? '你的回合' : '对手回合'}`;
}

function forfeit() {
    if (confirm('确定要认输吗？')) {
        send(MessageType.ForfeitReq, { room_id: currentRoom.id });
    }
}

function handleGameOver(payload) {
    const modal = document.getElementById('game-over-modal');
    const result = document.getElementById('game-result');
    
    if (payload.winner === currentUser.id) {
        result.textContent = '你赢了！';
        result.style.color = '#4ecca3';
    } else {
        result.textContent = '你输了！';
        result.style.color = '#e94560';
    }
    
    modal.classList.remove('hidden');
}

function backToLobby() {
    document.getElementById('game-over-modal').classList.add('hidden');
    currentGame = null;
    currentRoom = null;
    showPage('lobby-page');
    send(MessageType.RoomList, {});
    send(MessageType.UserStatsReq, { user_id: currentUser.id });
    send(MessageType.LeaderboardReq, { limit: 10 });
}

setInterval(() => {
    if (ws && ws.readyState === WebSocket.OPEN) {
        send(MessageType.Ping, {});
    }
}, 30000);

connect();
