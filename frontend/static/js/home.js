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
                        <!-- Post Creation Form -->
                        <div class="bg-white rounded-lg shadow-md p-4 mb-6">
                            <form id="create-post-form" class="space-y-4">
                                <div class="flex items-start space-x-4">
                                    <div class="flex-shrink-0">
                                        <div class="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center">
                                            👤
                                        </div>
                                    </div>
                                    <div class="flex-grow">
                                        <input type="text" 
                                            id="post-title" 
                                            name="title" 
                                            class="w-full p-3 border border-gray-200 rounded-lg mb-2" 
                                            placeholder="Title (optional)"
                                        >
                                        <textarea 
                                            id="post-content" 
                                            name="content" 
                                            class="w-full p-3 border border-gray-200 rounded-lg" 
                                            placeholder="What's on your mind?"
                                            rows="3"
                                            required
                                        ></textarea>
                                    </div>
                                </div>
                                <div class="flex items-center space-x-4">
                                    <input type="text" 
                                        id="post-categories" 
                                        name="categories" 
                                        class="flex-grow p-2 border border-gray-200 rounded-lg" 
                                        placeholder="Add categories (comma-separated)"
                                    >
                                    <label class="cursor-pointer flex items-center space-x-2 text-gray-600 hover:text-gray-800">
                                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                        </svg>
                                        <input type="file" 
                                            id="post-image" 
                                            name="image" 
                                            accept="image/*" 
                                            class="hidden"
                                        >
                                    </label>
                                </div>
                                <div id="image-preview" class="hidden mt-2">
                                    <img id="preview-img" src="" alt="Preview" class="max-h-64 rounded-lg">
                                    <button type="button" id="remove-image" class="mt-2 text-red-600 text-sm">Remove image</button>
                                </div>
                                <div class="flex justify-end">
                                    <button type="submit" class="bg-blue-500 text-white px-6 py-2 rounded-lg hover:bg-blue-600 transition-colors">
                                        Post
                                    </button>
                                </div>
                            </form>
                        </div>
                        <div id="feed">
                            <!-- Posts will be dynamically loaded here -->
                        </div>
                    </div>
                </div>
            </div>
        `;

        // Set up image preview functionality
        setupImagePreview();
        // Set up post form submission
        setupPostForm();
        // Load existing posts
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

            const data = await response.json();
            console.log('Loaded posts:', data);

            const posts = data.posts || [];
            const feedElement = document.getElementById('feed');
            
            if (!feedElement) {
                console.error('Feed element not found in the DOM');
                return;
            }

            feedElement.innerHTML = posts.map(post => `
                <div class="bg-white rounded-lg shadow-md p-4 mb-4">
                    <div class="flex items-center mb-4">
                        <div class="w-10 h-10 rounded-full bg-gray-300 flex items-center justify-center">
                            👤
                        </div>
                        <div class="ml-3">
                            <div class="flex items-baseline gap-2">
                                <span class="text-sm font-medium text-gray-700">@${post.user.nickname}</span>
                                <span class="text-xs text-gray-500">${new Date(post.createdAt).toLocaleString()}</span>
                            </div>
                            <h3 class="text-lg font-bold mt-1">${post.title || ''}</h3>
                        </div>
                    </div>
                    <p class="text-sm text-gray-600">${post.content}</p>
                    ${post.imageUrl ? `
                        <div class="mt-4">
                            <img src="${post.imageUrl}" alt="Post image" class="max-h-96 w-full object-cover rounded-lg">
                        </div>
                    ` : ''}
                    ${post.categories && post.categories.length > 0 ? `
                        <div class="mt-4 flex flex-wrap gap-2">
                            ${post.categories.map(cat => `
                                <span class="px-2 py-1 bg-gray-100 text-xs text-gray-600 rounded-full">${cat}</span>
                            `).join('')}
                        </div>
                    ` : ''}
                </div>
            `).join('');
        } catch (error) {
            console.error('Error loading feed:', error);
        }
    }

    function setupImagePreview() {
        const imageInput = document.getElementById('post-image');
        const previewContainer = document.getElementById('image-preview');
        const previewImg = document.getElementById('preview-img');
        const removeButton = document.getElementById('remove-image');

        imageInput.addEventListener('change', function(e) {
            const file = e.target.files[0];
            if (file) {
                const reader = new FileReader();
                reader.onload = function(e) {
                    previewImg.src = e.target.result;
                    previewContainer.classList.remove('hidden');
                }
                reader.readAsDataURL(file);
            }
        });

        removeButton.addEventListener('click', function() {
            imageInput.value = '';
            previewContainer.classList.add('hidden');
            previewImg.src = '';
        });
    }

    async function setupPostForm() {
        const form = document.getElementById('create-post-form');
        form.onsubmit = async (e) => {
            e.preventDefault();
            
            const formData = new FormData(form);
            console.log('Submitting post with data:', Object.fromEntries(formData)); // Debug log

            try {
                const response = await fetch('/api/posts/create', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    body: formData
                });

                if (!response.ok) {
                    throw new Error('Failed to create post');
                }

                const result = await response.json();
                console.log('Post creation response:', result); // Debug log

                // Clear form
                form.reset();
                
                // Reload feed to show new post
                await loadFeed();
                
                alert('Post created successfully!');
            } catch (error) {
                console.error('Error creating post:', error);
                alert('Failed to create post');
            }
        };
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
            console.error('Error during logout:', error);
        } finally {
            // Redirect to login page or perform any final actions
            window.location.href = '/login';
        }
    };

    render();
}