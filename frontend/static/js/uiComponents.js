import { utils } from './utils.js';

const uiComponents = {
    renderSignup() {
        return `
            <div class="min-h-screen flex items-center justify-center">
                <form id="signup-form" class="signup-form">
                    <h2>Sign Up</h2>
                    <div class="input-group">
                        <i class="fas fa-user"></i>
                        <input type="text" id="nickname" name="nickname" placeholder="Nickname" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-user"></i>
                        <input type="text" id="first_name" name="first_name" placeholder="First Name" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-user"></i>
                        <input type="text" id="last_name" name="last_name" placeholder="Last Name" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-envelope"></i>
                        <input type="email" id="email" name="email" placeholder="Email" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-lock"></i>
                        <input type="password" id="password" name="password" placeholder="Password" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-calendar"></i>
                        <input type="number" id="age" name="age" placeholder="Age" min="13" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-venus-mars"></i>
                        <select id="gender" name="gender" required>
                            <option value="">Select Gender</option>
                            <option value="male">Male</option>
                            <option value="female">Female</option>
                            <option value="other">Other</option>
                        </select>
                    </div>
                    <button type="submit">Sign Up</button>
                    <div id="error-info"></div>
                    <p>Already have an account? <a href="#login">Login</a></p>
                </form>
            </div>
        `;
    },

    renderLogin() {
        return `
            <div class="min-h-screen flex items-center justify-center">
                <form id="login-form" class="login-form">
                    <h2>Login</h2>
                    <div class="input-group">
                        <i class="fas fa-user"></i>
                        <input type="text" id="username" name="identifier" placeholder="Email or Nickname" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-lock"></i>
                        <input type="password" id="password" name="password" placeholder="Password" required>
                    </div>
                    <button type="submit">Login</button>
                    <div id="error-info"></div>
                    <p>Don't have an account? <a href="#signup">Sign Up</a></p>
                </form>
            </div>
        `;
    },

    renderHome(categories = [], user = {}, messages = []) {
        return `
            <style>
                .post-toggle-btn {
                    background-color: #3b82f6;
                    color: white;
                    padding: 0.5rem 1rem;
                    border-radius: 0.5rem;
                    border: none;
                    font-weight: 500;
                    cursor: pointer;
                    transition: background-color 0.2s;
                }
                .post-toggle-btn:hover {
                    background-color: #2563eb;
                }
                .create-post-btn {
                    background-color: #10b981;
                    color: white;
                    padding: 0.5rem 1rem;
                    border-radius: 0.5rem;
                    border: none;
                    font-weight: 500;
                    cursor: pointer;
                    transition: background-color 0.2s;
                    width: 100%;
                }
                .create-post-btn:hover {
                    background-color: #059669;
                }
                .success-message, .error-message {
                    position: fixed;
                    top: 1rem;
                    right: 1rem;
                    padding: 0.75rem 1.5rem;
                    border-radius: 0.5rem;
                    z-index: 1000;
                }
                .success-message {
                    background-color: #10b981;
                    color: white;
                }
                .error-message {
                    background-color: #ef4444;
                    color: white;
                }
                #post-form-container {
                    background-color: #f9fafb;
                    padding: 1.5rem;
                    border-radius: 0.75rem;
                    margin-bottom: 1rem;
                    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
                }
                .input-group {
                    display: flex;
                    align-items: center;
                    margin-bottom: 1rem;
                    background-color: white;
                    border: 1px solid #d1d5db;
                    border-radius: 0.5rem;
                    padding: 0.5rem;
                }
                .input-group i {
                    margin-right: 0.5rem;
                    color: #6b7280;
                }
                .input-group input, .input-group textarea, .input-group select {
                    flex: 1;
                    border: none;
                    outline: none;
                    font-size: 1rem;
                }
                #image-preview {
                    margin-top: 1rem;
                }
                #preview-image {
                    max-width: 100%;
                    border-radius: 0.5rem;
                }
                #remove-image-btn {
                    background-color: #ef4444;
                    color: white;
                    padding: 0.5rem 1rem;
                    border-radius: 0.5rem;
                    border: none;
                    margin-top: 0.5rem;
                    cursor: pointer;
                }
                #remove-image-btn:hover {
                    background-color: #dc2626;
                }
                .profile-section {
                    background-color: #f9fafb;
                    padding: 2rem;
                    border-radius: 0.75rem;
                    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
                    margin-bottom: 2rem;
                }
                .posts-grid {
                    display: grid;
                    gap: 1.5rem;
                }
                .post-card {
                    background-color: white;
                    padding: 1.5rem;
                    border-radius: 0.75rem;
                    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
                }
            </style>
            <header>
                <div class="header-title">Forum</div>
                <div class="header-right">
                    <button class="profile-icon"><i class="fas fa-user"></i></button>
                    <button id="logout-btn">Logout</button>
                </div>
            </header>
            <main>
                <aside id="category-sidebar">
                    <h2 class="text-lg font-bold mb-4">Categories</h2>
                    <table id="category-table">
                        <tbody>
                            <tr>
                                <td class="category-item" data-category-id="0">All</td>
                            </tr>
                            ${categories.length > 0 ? categories.map(category => `
                                <tr>
                                    <td class="category-item" data-category-id="${category.id}">${category.name}</td>
                                </tr>
                            `).join('') : `
                                <tr>
                                    <td class="text-gray-300">No categories available</td>
                                </tr>
                            `}
                        </tbody>
                    </table>
                </aside>
                <section id="feeds-section">
                    <div class="flex justify-end mb-4">
                        <button id="toggle-post-form-btn" class="post-toggle-btn">+ New Post</button>
                    </div>
                    <div id="post-form-container" class="hidden">
                        ${this.renderPostForm(categories)}
                    </div>
                    <div id="feed"></div>
                </section>
                <aside id="messaging-sidebar">
                    <h2 class="text-lg font-bold mb-4">Messages</h2>
                    <table id="messaging-table">
                        <tbody>
                            ${messages.length > 0 ? messages.map(message => `
                                <tr>
                                    <td class="message-item" data-message-id="${message.id}">
                                        <span class="font-medium">${message.sender_nickname || 'Unknown'}</span>: ${message.content.substring(0, 30)}${message.content.length > 30 ? '...' : ''}
                                    </td>
                                </tr>
                            `).join('') : `
                                <tr>
                                    <td class="text-gray-300">No messages available</td>
                                </tr>
                            `}
                        </tbody>
                    </table>
                </aside>
            </main>
            <footer>
                <p>© 2025 Forum. All rights reserved.</p>
            </footer>
        `;
    },

    renderProfilePage(user = {}, createdPosts = [], likedPosts = []) {
        return `
            <style>
                .profile-section {
                    background-color: #f9fafb;
                    padding: 2rem;
                    border-radius: 0.75rem;
                    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
                    margin-bottom: 2rem;
                }
                .posts-grid {
                    display: grid;
                    gap: 1.5rem;
                }
                .post-card {
                    background-color: white;
                    padding: 1.5rem;
                    border-radius: 0.75rem;
                    box-shadow: 0 1px 3px rgba(0,0,0,0.1);
                }
                .back-btn {
                    background-color: #3b82f6;
                    color: white;
                    padding: 0.5rem 1rem;
                    border-radius: 0.5rem;
                    border: none;
                    font-weight: 500;
                    cursor: pointer;
                    transition: background-color 0.2s;
                    margin-bottom: 1rem;
                }
                .back-btn:hover {
                    background-color: #2563eb;
                }
            </style>
            <header>
                <div class="header-title">Forum</div>
                <div class="header-right">
                    <button class="profile-icon"><i class="fas fa-user"></i></button>
                    <button id="logout-btn">Logout</button>
                </div>
            </header>
            <main class="container mx-auto px-4 py-8">
                <button id="back-btn" class="back-btn">Back to Home</button>
                <section class="profile-section">
                    <h2 class="text-2xl font-bold mb-4">Profile</h2>
                    <p><strong>Nickname:</strong> ${user.nickname || 'Unknown'}</p>
                    <p><strong>Email:</strong> ${user.email || 'Not provided'}</p>
                    <p><strong>Joined:</strong> ${user.created_at ? utils.formatDate(user.created_at) : 'Not provided'}</p>
                </section>
                <section class="posts-grid">
                    <h3 class="text-xl font-bold mb-4">Created Posts</h3>
                    ${createdPosts.length ? createdPosts.map(post => this.renderPost(post)).join('') : '<p>No posts created yet.</p>'}
                </section>
                <section class="posts-grid">
                    <h3 class="text-xl font-bold mb-4">Liked Posts</h3>
                    ${likedPosts.length ? likedPosts.map(post => this.renderPost(post)).join('') : '<p>No posts liked yet.</p>'}
                </section>
            </main>
            <footer>
                <p>© 2025 Forum. All rights reserved.</p>
            </footer>
        `;
    },

    renderPostForm(categories = []) {
        return `
            <form id="create-post-form" method="POST" enctype="multipart/form-data">
                <div class="input-group">
                    <i class="fas fa-heading"></i>
                    <input type="text" id="post-title" name="title" placeholder="Title (optional)">
                </div>
                <div class="input-group">
                    <i class="fas fa-comment"></i>
                    <textarea id="post-content" name="content" placeholder="What's on your mind?" rows="4" required></textarea>
                </div>
                <div class="input-group">
                    <i class="fas fa-list"></i>
                    <select id="post-category" name="category_id" required>
                        <option value="">Select Category</option>
                        ${categories.map(category => `<option value="${category.id}">${category.name}</option>`).join('')}
                    </select>
                </div>
                <div class="input-group">
                    <i class="fas fa-image"></i>
                    <input type="file" id="post-image" name="image" accept="image/*">
                </div>
                <div id="image-preview" class="hidden">
                    <img id="preview-image" src="" alt="Image preview">
                    <button type="button" id="remove-image-btn">Remove Image</button>
                </div>
                <button type="submit" class="create-post-btn">Post</button>
            </form>
        `;
    },

    renderPost(post) {
        const comments = post.comments || [];
        return `
            <div class="post-card" data-post-id="${post.id}">
                <h3 class${post.title ? '' : 'hidden'}>${post.title || ''}</h3>
                <p>${post.content}</p>
                ${post.image_url ? `<img src="${post.image_url}" alt="Post image" class="post-image">` : ''}
                <div class="post-meta">
                    <span>By ${post.author_nickname || 'Unknown'}</span>
                    <span>${utils.formatDate(post.created_at)}</span>
                </div>
                <div class="post-actions">
                    <button class="like-btn" data-post-id="${post.id}"><i class="fas fa-thumbs-up"></i> ${post.reactions?.likes || 0}</button>
                    <button class="dislike-btn" data-post-id="${post.id}"><i class="fas fa-thumbs-down"></i> ${post.reactions?.dislikes || 0}</button>
                    <button class="comment-btn" data-post-id="${post.id}"><i class="fas fa-comment"></i> ${comments.length}</button>
                </div>
                <div class="comments-section hidden" data-post-id="${post.id}">
                    <div class="comments-list">
                        ${comments.length ? comments.map(comment => this.renderComment(comment, post.current_user_id)).join('') : '<p>No comments yet.</p>'}
                    </div>
                    <button class="show-comment-form-btn" data-post-id="${post.id}">Add Comment</button>
                    <form class="comment-form hidden" data-post-id="${post.id}">
                        <div class="input-group">
                            <i class="fas fa-comment"></i>
                            <input type="text" name="content" placeholder="Add a comment..." required>
                        </div>
                        <button type="submit">Post</button>
                    </form>
                </div>
            </div>
        `;
    },

    renderFeed(posts = []) {
        if (!posts.length) {
            return `<div id="feeds-table" class="no-posts-msg">No Feeds Available</div>`;
        }
        return `
            <div id="feeds-table">
                ${posts
                    .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
                    .map(post => this.renderPost(post)).join('')}
            </div>
        `;
    },

    renderComment(comment, currentUserId) {
        return `
            <div class="comment" data-comment-id="${comment.id}">
                <div class="post-meta">
                    <span>@${comment.user?.nickname || 'Unknown'}</span>
                    <span>${utils.formatDate(comment.created_at)}</span>
                    ${comment.user?.id === currentUserId ? `
                        <button class="edit-comment-btn">Edit</button>
                        <button class="delete-comment-btn">Delete</button>
                    ` : ''}
                </div>
                <div class="comment-content">
                    <p>${comment.content}</p>
                </div>
                <form class="edit-comment-form hidden" data-comment-id="${comment.id}">
                    <div class="input-group">
                        <i class="fas fa-comment"></i>
                        <input type="text" value="${comment.content}" name="content" required>
                    </div>
                    <button type="submit">Save</button>
                    <button type="button" class="cancel-edit-btn">Cancel</button>
                </form>
            </div>
        `;
    },

    renderMessageView(message) {
        return `
            <div class="message-view" style="position: fixed; top: 20%; left: 50%; transform: translateX(-50%); background: white; padding: 2rem; border-radius: 0.75rem; box-shadow: 0 4px 12px rgba(0,0,0,0.1); z-index: 1000;">
                <h3>Message from @${message.sender_nickname || 'Unknown'}</h3>
                <p>${message.content}</p>
                <span>${utils.formatDate(message.created_at)}</span>
                <button class="close-message-btn" style="background: #3b82f6; color: white; padding: 0.5rem 1rem; border-radius: 0.5rem; border: none; margin-top: 1rem;">Close</button>
            </div>
        `;
    }
};

export { uiComponents };