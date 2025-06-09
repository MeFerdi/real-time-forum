function renderHome() {
    let wsClient = null;

    function render() {
        // Initialize WebSocket with user ID from token
        if (!window.wsClient) {
            const token = localStorage.getItem('token');
            if (token) {
                try {
                    const payload = JSON.parse(atob(token.split('.')[1]));
                    wsClient = new MessageWebSocket(payload.user_id);
                    window.wsClient = wsClient; // Make it globally accessible
                    console.log('WebSocket client initialized with user ID:', payload.user_id);
                } catch (e) {
                    console.error('Failed to initialize WebSocket:', e);
                }
            }
        } else {
            wsClient = window.wsClient;
        }

        document.getElementById('app').innerHTML = `
            <div class="min-h-screen bg-gray-100">
                <header class="bg-gray-800 text-gray-300 shadow-lg sticky top-0 z-50">
                    <div class="max-w-6xl mx-auto px-4">
                        <div class="flex items-center justify-between h-16">
                            <!-- Logo -->
                            <div class="text-2xl font-bold flex items-center">
                                <span>RealTime Forum</span>
                            </div>

                            <!-- Navbar -->
                            <nav class="flex space-x-6">
                                <button class="nav-btn text-gray-300 hover:text-gray-100">
                                    🏠 Home
                                </button>
                                <button class="nav-btn text-gray-300 hover:text-gray-100">
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

    async function handleReaction(postId, reactionType) {
        try {
            const response = await fetch(`/api/posts/${postId}/react`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({ reactionType })
            });

            if (!response.ok) {
                throw new Error('Failed to update reaction');
            }

            const result = await response.json();
            
            // Update UI
            const postElement = document.querySelector(`[data-post-id="${postId}"]`);
            if (postElement) {
                const likeCount = postElement.querySelector('.like-count');
                const dislikeCount = postElement.querySelector('.dislike-count');
                const likeBtn = postElement.querySelector('.like-btn');
                const dislikeBtn = postElement.querySelector('.dislike-btn');

                if (likeCount) likeCount.textContent = result.likeCount;
                if (dislikeCount) dislikeCount.textContent = result.dislikeCount;

                // Reset active states
                likeBtn.classList.remove('text-blue-600', 'bg-blue-50');
                dislikeBtn.classList.remove('text-red-600', 'bg-red-50');

                // Set active state based on current reaction
                if (result.userReaction === 'like') {
                    likeBtn.classList.add('text-blue-600', 'bg-blue-50');
                } else if (result.userReaction === 'dislike') {
                    dislikeBtn.classList.add('text-red-600', 'bg-red-50');
                }
            }
        } catch (error) {
            console.error('Error handling reaction:', error);
            alert('Failed to update reaction');
        }
    }

    async function setupReactionHandlers() {
        document.querySelectorAll('.reaction-btn').forEach(btn => {
            btn.onclick = async function() {
                const postId = this.closest('[data-post-id]').dataset.postId;
                const currentReaction = this.classList.contains('like-btn') ? 'like' : 'dislike';
                const isActive = this.classList.contains(currentReaction === 'like' ? 'text-blue-600' : 'text-red-600');

                // If already active, remove reaction, otherwise add new reaction
                await handleReaction(postId, isActive ? '' : currentReaction);
            };
        });
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

            // Generate the HTML for posts
            const postHtml = posts.map(post => `
                <div class="bg-white rounded-lg shadow-md p-4 mb-4" data-post-id="${post.id}">
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
                    
                    <!-- Post Actions -->
                    <div class="mt-4 border-t pt-4">
                        <div class="flex justify-between items-center mb-4">
                            <button class="reaction-btn like-btn flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${post.userReaction === 'like' ? 'text-blue-600 bg-blue-50' : 'text-gray-600 hover:bg-gray-50'}">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path d="M2 10.5a1.5 1.5 0 113 0v6a1.5 1.5 0 01-3 0v-6zM6 10.333v5.43a2 2 0 001.106 1.79l.05.025A4 4 0 008.943 18h5.416a2 2 0 001.962-1.608l1.2-6A2 2 0 0015.56 8H12V4a2 2 0 00-2-2 1 1 0 00-1 1v.667a4 4 0 01-.8 2.4L6.8 7.933a4 4 0 00-.8 2.4z" />
                                </svg>
                                <span class="like-count">${post.likeCount || 0}</span>
                            </button>

                            <button class="reaction-btn comment-btn flex items-center gap-2 px-4 py-2 rounded-lg text-gray-600 hover:bg-gray-50 transition-colors">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path fill-rule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L2 17l1.338-3.123C2.493 12.767 2 11.434 2 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM7 9H5v2h2V9zm8 0h-2v2h2V9zM9 9h2v2H9V9z" clip-rule="evenodd" />
                                </svg>
                                <span>${post.comments ? post.comments.length : 0}</span>
                            </button>

                            <button class="reaction-btn dislike-btn flex items-center gap-2 px-4 py-2 rounded-lg transition-colors ${post.userReaction === 'dislike' ? 'text-red-600 bg-red-50' : 'text-gray-600 hover:bg-gray-50'}">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path d="M18 9.5a1.5 1.5 0 11-3 0v-6a1.5 1.5 0 013 0v6zM14 9.667v-5.43a2 2 0 00-1.105-1.79l-.05-.025A4 4 0 0011.055 2H5.64a2 2 0 00-1.962 1.608l-1.2 6A2 2 0 004.44 12H8v4a2 2 0 002 2 1 1 0 001-1v-.667a4 4 0 01.8-2.4l1.4-1.866a4 4 0 00.8-2.4z" />
                                </svg>
                                <span class="dislike-count">${post.dislikeCount || 0}</span>
                            </button>
                        </div>

                        <!-- Comments Section -->
                        <div class="mt-4">
                            <div class="mb-4">
                                <form class="comment-form" data-post-id="${post.id}">
                                    <div class="flex gap-2">
                                        <input type="text" 
                                            class="flex-grow p-2 border rounded-lg" 
                                            placeholder="Write a comment..."
                                            name="content"
                                            required
                                        >
                                        <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600">
                                            Comment
                                        </button>
                                    </div>
                                </form>
                            </div>
                            <div class="space-y-3 comments-container">
                                ${post.comments ? post.comments.map(comment => `
                                    <div class="bg-gray-50 p-3 rounded-lg" data-comment-id="${comment.id}">
                                        <div class="flex items-center justify-between mb-1">
                                            <div class="flex items-center gap-2">
                                                <span class="font-medium text-sm">@${comment.user.nickname}</span>
                                                <span class="text-xs text-gray-500">${new Date(comment.createdAt).toLocaleString()}</span>
                                            </div>
                                            ${comment.user.id === JSON.parse(atob(localStorage.getItem('token').split('.')[1])).user_id ? `
                                                <div class="flex items-center gap-2">
                                                    <button class="edit-comment-btn text-xs text-blue-600 hover:text-blue-800">Edit</button>
                                                    <button class="delete-comment-btn text-xs text-red-600 hover:text-red-800">Delete</button>
                                                </div>
                                            ` : ''}
                                        </div>
                                        <div class="comment-content">
                                            <p class="text-sm text-gray-700">${comment.content}</p>
                                        </div>
                                        <form class="edit-comment-form hidden mt-2">
                                            <div class="flex gap-2">
                                                <input type="text" class="flex-grow p-2 border rounded-lg text-sm" value="${comment.content}">
                                                <button type="submit" class="bg-blue-500 text-white px-3 py-1 rounded-lg text-sm hover:bg-blue-600">Save</button>
                                                <button type="button" class="cancel-edit-btn bg-gray-300 text-gray-700 px-3 py-1 rounded-lg text-sm hover:bg-gray-400">Cancel</button>
                                            </div>
                                        </form>
                                    </div>
                                `).join('') : ''}
                            </div>
                        </div>
                    </div>
                </div>
            `).join('');
            
            // Set the post HTML
            feedElement.innerHTML = postHtml;

            // Set up forms and actions
            setupCommentForms();
            setupCommentActions();
            setupReactionHandlers();
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

    // Set up comment form handlers after loading posts
    function setupCommentForms() {
        console.log('Setting up comment forms');
        
        // Remove any existing event listeners
        document.querySelectorAll('.comment-form').forEach(form => {
            const newForm = form.cloneNode(true);
            form.parentNode.replaceChild(newForm, form);
        });

        // Attach new event listeners
        document.querySelectorAll('.comment-form').forEach(form => {
            const postId = form.getAttribute('data-post-id');
            console.log('Attaching submit handler to comment form:', postId);

            form.addEventListener('submit', async (e) => {
                e.preventDefault();
                console.log('Comment form submitted');

                const contentInput = form.querySelector('input[name="content"]');
                if (!contentInput) {
                    console.error('No content input found in form');
                    return;
                }

                const content = contentInput.value.trim();
                if (!content) {
                    alert('Please write a comment');
                    return;
                }

                console.log(`Submitting comment for post ${postId}:`, content);
                const submitButton = form.querySelector('button[type="submit"]');
                submitButton.disabled = true;

                try {
                    const response = await fetch(`/api/posts/${postId}/comments/`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        },
                        body: JSON.stringify({ content })
                    });

                    let result;
                    const responseText = await response.text();
                    console.log('Server response:', responseText);

                    try {
                        result = JSON.parse(responseText);
                    } catch (e) {
                        console.error('Failed to parse response as JSON:', e);
                    }

                    if (!response.ok) {
                        throw new Error(result?.message || 'Failed to add comment');
                    }

                    console.log('Comment added successfully:', result);
                    form.reset();

                    // If WebSocket is not connected and working, reload the feed
                    if (!window.wsClient?.connection?.readyState === WebSocket.OPEN) {
                        console.log('WebSocket not connected, reloading feed');
                        await loadFeed();
                    } else {
                        console.log('WebSocket connected, waiting for real-time update');
                    }
                } catch (error) {
                    console.error('Error adding comment:', error);
                    alert('Failed to add comment: ' + error.message);
                } finally {
                    submitButton.disabled = false;
                }
            });
        });
    }

    // Set up comment edit and delete handlers
    function setupCommentActions() {
        document.querySelectorAll('.edit-comment-btn').forEach(btn => {
            btn.onclick = function() {
                const commentDiv = btn.closest('[data-comment-id]');
                const content = commentDiv.querySelector('.comment-content');
                const form = commentDiv.querySelector('.edit-comment-form');
                const input = form.querySelector('input');
                
                content.classList.add('hidden');
                form.classList.remove('hidden');
                input.value = input.value.trim();
                input.focus();
            };
        });

        document.querySelectorAll('.cancel-edit-btn').forEach(btn => {
            btn.onclick = function() {
                const commentDiv = btn.closest('[data-comment-id]');
                const content = commentDiv.querySelector('.comment-content');
                const form = commentDiv.querySelector('.edit-comment-form');
                
                content.classList.remove('hidden');
                form.classList.add('hidden');
            };
        });

        document.querySelectorAll('.edit-comment-form').forEach(form => {
            form.onsubmit = async function(e) {
                e.preventDefault();
                const commentDiv = form.closest('[data-comment-id]');
                const commentId = commentDiv.dataset.commentId;
                const content = form.querySelector('input').value.trim();

                try {
                    const response = await fetch(`/api/posts/comments/${commentId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        },
                        body: JSON.stringify({ content })
                    });

                    if (!response.ok) {
                        throw new Error('Failed to update comment');
                    }

                    // Update will be handled by WebSocket
                } catch (error) {
                    console.error('Error updating comment:', error);
                    alert('Failed to update comment');
                }
            };
        });

        document.querySelectorAll('.delete-comment-btn').forEach(btn => {
            btn.onclick = async function() {
                if (!confirm('Are you sure you want to delete this comment?')) {
                    return;
                }

                const commentDiv = btn.closest('[data-comment-id]');
                const commentId = commentDiv.dataset.commentId;

                try {
                    const response = await fetch(`/api/posts/comments/${commentId}`, {
                        method: 'DELETE',
                        headers: {
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        }
                    });

                    if (!response.ok) {
                        throw new Error('Failed to delete comment');
                    }

                    // Deletion will be handled by WebSocket
                } catch (error) {
                    console.error('Error deleting comment:', error);
                    alert('Failed to delete comment');
                }
            };
        });
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