class AuthService {
    constructor() {
        this.token = localStorage.getItem('token');
        this.userId = null;
        this.nickname = '';
        this.ws = null;
    }

    async signup(userData) {
        // Ensure all required fields are present and in snake_case
        const payload = {
            email: userData.email,
            nickname: userData.nickname,
            password: userData.password,
            first_name: userData.first_name,
            last_name: userData.last_name,
            age: userData.age,
            gender: userData.gender
        };
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
            
            const data = await response.json();
            
            if (response.ok) {
                this.token = data.token;
                localStorage.setItem('token', this.token);
                const payload = JSON.parse(atob(this.token.split('.')[1]));
                this.userId = payload.user_id;
                this.nickname = payload.nickname;
                this.ws = new MessageWebSocket(this.userId);
                return true;
            } else {
                throw new Error(data.message || 'Signup failed');
            }
        } catch (error) {
            throw error;
        }
    }

    async login(credentials) {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(credentials),
            });
            
            const data = await response.json();
            
            if (response.ok) {
                this.token = data.token;
                localStorage.setItem('token', this.token);
                const payload = JSON.parse(atob(this.token.split('.')[1]));
                this.userId = payload.user_id;
                this.nickname = payload.nickname;
                this.ws = new MessageWebSocket(this.userId);
                return true;
            } else {
                throw new Error(data.message || 'Login failed');
            }
        } catch (error) {
            throw error;
        }
    }

    logout() {
        localStorage.removeItem('token');
        this.token = null;
        this.userId = null;
        this.nickname = '';
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }

    isAuthenticated() {
        return !!this.token;
    }

    getUserInfo() {
        return {
            userId: this.userId,
            nickname: this.nickname
        };
    }
}

// Remove export keyword to avoid ES module syntax error
const authService = new AuthService();

// Minimal JS for signup and login

function renderSignup() {
    document.getElementById('app').innerHTML = `
        <div class="flex items-center justify-center min-h-screen">
            <form id="signup-form" class="bg-white p-8 rounded shadow-md w-full max-w-md space-y-4">
                <h2 class="text-2xl font-bold mb-4 text-center">Sign Up</h2>
                <input type="text" id="nickname" placeholder="Nickname" class="w-full p-2 border rounded" required>
                <input type="text" id="first_name" placeholder="First Name" class="w-full p-2 border rounded" required>
                <input type="text" id="last_name" placeholder="Last Name" class="w-full p-2 border rounded" required>
                <input type="email" id="email" placeholder="Email" class="w-full p-2 border rounded" required>
                <input type="password" id="password" placeholder="Password" class="w-full p-2 border rounded" required>
                <input type="number" id="age" placeholder="Age" class="w-full p-2 border rounded" required>
                <select id="gender" class="w-full p-2 border rounded" required>
                    <option value="">Select Gender</option>
                    <option value="male">Male</option>
                    <option value="female">Female</option>
                    <option value="other">Other</option>
                </select>
                <button type="submit" class="w-full bg-blue-500 text-white p-2 rounded">Sign Up</button>
                <div id="signup-error" class="text-red-500 text-center"></div>
                <p class="text-center">Already have an account? <a href="#login" class="text-blue-500">Login</a></p>
            </form>
        </div>
    `;
    document.getElementById('signup-form').onsubmit = async function(e) {
        e.preventDefault();
        const password = document.getElementById('password').value;
        const errorEl = document.getElementById('signup-error');
        if (password.length < 8) {
            errorEl.textContent = 'Password must be at least 8 characters.';
            return;
        }
        errorEl.textContent = '';
        const data = {
            nickname: document.getElementById('nickname').value,
            first_name: document.getElementById('first_name').value,
            last_name: document.getElementById('last_name').value,
            email: document.getElementById('email').value,
            password: password,
            age: parseInt(document.getElementById('age').value, 10),
            gender: document.getElementById('gender').value
        };
        try {
            const res = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            if (res.ok) {
                renderLogin();
            } else {
                const err = await res.json();
                errorEl.textContent = err.error || 'Signup failed';
            }
        } catch (e) {
            errorEl.textContent = 'Network error';
        }
    };
    document.getElementById('password').addEventListener('input', function() {
        const errorEl = document.getElementById('signup-error');
        if (this.value.length >= 8) {
            errorEl.textContent = '';
        }
    });
    document.querySelector('a[href="#login"]').onclick = function(e) {
        e.preventDefault();
        renderLogin();
    };
}

function renderLogin() {
    document.getElementById('app').innerHTML = `
        <div class="flex items-center justify-center min-h-screen">
            <form id="login-form" class="bg-white p-8 rounded shadow-md w-full max-w-md space-y-4">
                <h2 class="text-2xl font-bold mb-4 text-center">Login</h2>
                <input type="text" id="identifier" placeholder="Email or Nickname" class="w-full p-2 border rounded" required>
                <input type="password" id="password" placeholder="Password" class="w-full p-2 border rounded" required>
                <button type="submit" class="w-full bg-blue-500 text-white p-2 rounded">Login</button>
                <div id="login-error" class="text-red-500 text-center"></div>
                <p class="text-center">Don't have an account? <a href="#signup" class="text-blue-500">Sign Up</a></p>
            </form>
        </div>
    `;
    document.getElementById('login-form').onsubmit = async function(e) {
        e.preventDefault();
        const data = {
            identifier: document.getElementById('identifier').value,
            password: document.getElementById('password').value
        };
        try {
            const res = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data)
            });
            if (res.ok) {
                // On successful login, render the Home page
                if (typeof renderHome === 'function') {
                    renderHome();
                } else if (window.renderHome) {
                    window.renderHome();
                }
            } else {
                const err = await res.json();
                document.getElementById('login-error').textContent = err.error || 'Login failed';
            }
        } catch (e) {
            document.getElementById('login-error').textContent = 'Network error';
        }
    };
    document.querySelector('a[href="#signup"]').onclick = function(e) {
        e.preventDefault();
        renderSignup();
    };
}

// Initial render
if (window.location.hash === '#login') {
    renderLogin();
} else {
    renderSignup();
}