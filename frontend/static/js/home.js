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
                                <button class="nav-btn text-gray-300 hover:text-gray-100" onclick="renderCreatePost()">
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
                        <div id="feed">
                            <!-- Posts will be dynamically loaded here -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        loadFeed();
    }

    async function loadFeed() {
        try {
            const response = await fetch('/api/posts', {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (!response.ok) {
                throw new Error('Failed to load feed');
            }

            const posts = await response.json();
            const feedElement = document.getElementById('feed');
            feedElement.innerHTML = posts.map(post => `
                <div class="bg-white rounded-lg shadow-md p-4 mb-4">
                    <h3 class="text-lg font-bold">${post.title}</h3>
                    <p class="text-sm text-gray-600">${post.content}</p>
                    <p class="text-xs text-gray-500">Categories: ${post.categories.join(', ')}</p>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error loading feed:', error);
        }
    }

    window.renderCreatePost = function() {
        document.getElementById('main-content').innerHTML = `
            <div class="bg-white rounded-lg shadow-md p-6">
                <h2 class="text-xl font-bold mb-4">Create a Post</h2>
                <form id="create-post-form">
                    <div class="mb-4">
                        <label for="post-title" class="block text-sm font-medium text-gray-700">Title</label>
                        <input type="text" id="post-title" name="title" class="mt-1 block w-full p-2 border border-gray-300 rounded-md" required>
                    </div>
                    <div class="mb-4">
                        <label for="post-content" class="block text-sm font-medium text-gray-700">Content</label>
                        <textarea id="post-content" name="content" rows="4" class="mt-1 block w-full p-2 border border-gray-300 rounded-md" required></textarea>
                    </div>
                    <div class="mb-4">
                        <label for="post-categories" class="block text-sm font-medium text-gray-700">Categories</label>
                        <input type="text" id="post-categories" name="categories" placeholder="Comma-separated categories" class="mt-1 block w-full p-2 border border-gray-300 rounded-md">
                    </div>
                    <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600">Submit</button>
                </form>
            </div>
        `;

        document.getElementById('create-post-form').addEventListener('submit', async function(event) {
            event.preventDefault();

            const title = document.getElementById('post-title').value;
            const content = document.getElementById('post-content').value;
            const categories = document.getElementById('post-categories').value.split(',').map(cat => cat.trim());

            try {
                const response = await fetch('/api/posts', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    body: JSON.stringify({
                        title,
                        content,
                        categories
                    })
                });

                if (!response.ok) {
                    throw new Error('Failed to create post');
                }

                alert('Post created successfully!');
                loadFeed();
            } catch (error) {
                console.error('Error creating post:', error);
                alert('Failed to create post. Please try again.');
            }
        });
    };

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