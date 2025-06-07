function renderHome() {
    function render() {
        document.getElementById('app').innerHTML = `
            <div class="min-h-screen bg-gray-100">
                <header class="bg-gray-800 text-gray-300 shadow-lg sticky top-0 z-50">
                    <div class="max-w-6xl mx-auto px-4">
                        <div class="flex items-center justify-between h-16">
                            <!-- Logo -->
                            <div class="text-2xl font-bold flex items-center">
                                <span class="mr-2">🌍</span>
                                <span>Ubuntu Connect</span>
                            </div>

                            <!-- Navbar -->
                            <nav class="flex space-x-6">
                                <button class="nav-btn text-gray-300 hover:text-gray-100" onclick="switchView('home')">
                                    🏠 Home
                                </button>
                                <button class="nav-btn text-gray-300 hover:text-gray-100" onclick="switchView('profile')">
                                    👤 Profile
                                </button>
                                <button class="nav-btn text-gray-300 hover:text-gray-100" onclick="switchView('createPost')">
                                    ✍️ Create Post
                                </button>
                            </nav>

                            <!-- Logout -->
                            <button onclick="logout()" class="ml-2 px-3 py-1 bg-red-500 hover:bg-red-600 rounded-full text-sm transition-colors">
                                Logout
                            </button>
                        </div>
                    </div>
                </header>

                <!-- Main Content -->
                <div class="max-w-6xl mx-auto px-4 py-6">
                    <div id="main-content">
                        <!-- Content will be dynamically loaded here -->
                    </div>
                </div>
            </div>
        `;
    }

    // Global functions
    window.logout = async function() {
        try {
            // Clear local storage and any auth tokens
            localStorage.removeItem('token');
            
            // Close WebSocket connection if it exists
            if (window.authService && window.authService.ws) {
                window.authService.ws.close();
            }

            // Call the backend logout endpoint if needed
            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                console.error('Logout failed:', response.statusText);
            }
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            // Always redirect to login page and clean up state
            if (window.authService) {
                window.authService.token = null;
                window.authService.userId = null;
                window.authService.nickname = '';
            }
            window.location.hash = '#login';
            if (typeof renderLogin === 'function') {
                renderLogin();
            }
        }
    };

    // Initial render
    render();
}

// Make renderHome globally accessible
window.renderHome = renderHome;