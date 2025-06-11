import { utils } from './utils.js';

const uiComponents = {
    renderSignup() {
        return `
            <div class="min-h-screen">
                <form id="signup-form" class="signup-form">
                    <h2>Sign Up</h2>
                    <div class="input-group">
                        <i class="fas fa-user"></i>
                        <input type="text" id="nickname" name="nickname" placeholder="Nickname" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-id-card"></i>
                        <input type="text" id="first_name" name="first_name" placeholder="First Name" required>
                    </div>
                    <div class="input-group">
                        <i class="fas fa-id-card"></i>
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
                        <input type="number" id="age" name="age" placeholder="Age" required>
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
                    <button type="submit"><i class="fas fa-paper-plane"></i> Sign Up</button>
                    <div id="error-info"></div>
                    <p class="text-center">Already have an account? <a href="#login">Login</a></p>
                </form>
            </div>
        `;
    },

    renderLogin() {
        return `
            <div class="min-h-screen">
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
                    <button type="submit"><i class="fas fa-sign-in-alt"></i> Login</button>
                    <div id="error-info"></div>
                    <p class="text-center">Don't have an account? <a href="#signup">Sign Up</a></p>
                </form>
            </div>
        `;
    },

    renderHome(categories = [], user = {}) {
    return `
        <div class="min-h-screen flex">
            <!-- Category Sidebar -->
            <aside id="category-sidebar" class="w-64 bg-gray-900 p-4">
                <h2 class="text-lg font-bold mb-4 text-neon-cyan">Categories</h2>
                <table id="category-table" class="w-full">
                    <tbody>
                        <tr>
                            <td class="category-item p-2 cursor-pointer hover:bg-gray-800" data-category-id="0">All</td>
                        </tr>
                        ${categories.length > 0 ? categories.map(category => `
                            <tr>
                                <td class="category-item p-2 cursor-pointer hover:bg-gray-800" data-category-id="${category.id}">${category.name}</td>
                            </tr>
                        `).join('') : `
                            <tr>
                                <td class="p-2 text-gray-400">No categories available</td>
                            </tr>
                        `}
                    </tbody>
                </table>
            </aside>
            <!-- Main Content -->
            <div class="flex-1">
                <header>
                    <div class="flex items-center justify-between">
                        <div class="text-2xl">
                            <span>RealTime Forum</span>
                        </div>
                        <nav class="flex gap-4">
                            <button class="nav-btn" data-nav="home">🏠 Home</button>
                        </nav>
                        <div class="relative">
                            <button id="profile-btn" class="profile-icon" title="Profile">
                                <i class="fas fa-user"></i>
                            </button>
                            <div id="profile-dropdown" class="profile-dropdown hidden">
                                <div class="profile-info">
                                    <div class="profile-avatar mb-2"><i class="fas fa-user-circle fa-2x"></i></div>
                                    <div><strong>Nickname:</strong> <span id="profile-nickname">${user.nickname || 'Guest'}</span></div>
                                    <div><strong>Email:</strong> <span id="profile-email">${user.email || 'guest@example.com'}</span></div>
                                    <div><strong>Joined:</strong> <span id="profile-joined">${user.created_at ? utils.formatDate(user.created_at) : '2024-01-01'}</span></div>
                                </div>
                            </div>
                        </div>
                        <button id="logout-btn">Logout</button>
                    </div>
                </header>
                <div id="main-content">
                    <div class="flex justify-end mb-4">
                        <button id="toggle-post-form-btn" class="post-toggle-btn">
                            <i class="fas fa-plus"></i> Post
                        </button>
                    </div>
                    <div id="post-form-container" class="hidden">
                        ${this.renderPostForm(categories)}
                    </div>
                    <div id="feed"></div>
                </div>
            </div>
        </div>
    `;
},

    renderPostForm(categories = []) {
        return `
            <div>
                <form id="create-post-form" class="space-y-4" method="POST" enctype="multipart/form-data">
                    <div class="flex items-start">
                        <div class="flex-shrink-0">
                            <div class="avatar">👤</div>
                        </div>
                        <div class="flex-grow">
                            <input type="text" id="post-title" name="title" class="w-full" placeholder="Title (optional)">
                            <textarea id="post-content" name="content" class="w-full" placeholder="What's on your mind?" rows="3" required></textarea>
                        </div>
                    </div>
                    <div class="flex items-center gap-2">
                        <select id="post-category" name="category_id" class="flex-grow bg-gray-900 text-light border border-gray-700 rounded p-2" required>
                            <option value="">Select Category</option>
                            ${categories.map(category => `
                                <option value="${category.id}">${category.name}</option>
                            `).join('')}
                        </select>
                        <label class="cursor-pointer">
                            <i class="fas fa-camera"></i>
                            <input type="file" id="post-image" name="image" accept="image/*" class="hidden">
                        </label>
                    </div>
                    <div id="image-preview" class="hidden">
                        <img id="preview-image" src="" alt="Image preview">
                        <button type="button" id="remove-image-btn"><i class="fas fa-trash"></i> Remove image</button>
                    </div>
                    <div class="flex justify-end">
                        <button type="submit"><i class="fas fa-paper-plane"></i> Post</button>
                    </div>
                </form>
            </div>
        `;
    },

    renderPost(post) {
        return `
            <div data-post-id="${post.id}" class="post-card">
                <div class="flex items-center">
                    <div class="avatar">
                        <span>👤</span>
                    </div>
                    <div>
                        <div class="flex items-center gap-2">
                            <span class="text-sm">@${post.user.nickname}</span>
                            <span class="text-xs">${utils.formatDate(post.created_at)}</span>
                            ${post.category ? `<span class="category-tag">${post.category.name}</span>` : ''}
                        </div>
                        <h3>${post.title || ''}</h3>
                    </div>
                </div>
                <p>${post.content}</p>
                ${post.image_url ? `
                    <div>
                        <img src="${post.image_url}" alt="Post image" class="max-h-100">
                    </div>
                ` : ''}
                <div class="post-actions flex gap-3 mt-2 justify-end">
                    <button class="reaction-btn like-btn ${post.user_reaction === 'like' ? 'text-neon-cyan' : 'text-gray-400'}" data-post-id="${post.id}">
                        <i class="fas fa-thumbs-up"></i>
                        <span class="like-count">${post.like_count || 0}</span>
                    </button>
                    <button class="reaction-btn dislike-btn ${post.user_reaction === 'dislike' ? 'text-neon-pink' : 'text-gray-400'}" data-post-id="${post.id}">
                        <i class="fas fa-thumbs-down"></i>
                        <span class="dislike-count">${post.dislike_count || 0}</span>
                    </button>
                    <button class="reaction-btn comment-btn text-gray-400" data-post-id="${post.id}">
                        <i class="fas fa-comment"></i>
                        <span>${post.comments ? post.comments.length : 0}</span>
                    </button>
                </div>
                <div class="comments-section hidden">
                    <form class="comment-form" data-post-id="${post.id}">
                        <div class="flex">
                            <input type="text" placeholder="Write a comment..." name="content" required>
                            <button type="submit"><i class="fas fa-paper-plane"></i> Comment</button>
                        </div>
                    </form>
                    <div class="comments-container">
                        ${post.comments ? post.comments.map(comment => this.renderComment(comment, post.user.id)).join('') : ''}
                    </div>
                </div>
            </div>
        `;
    },

    renderComment(comment, currentUserId) {
        return `
            <div data-comment-id="${comment.id}">
                <div class="flex">
                    <div class="flex">
                        <span>@${comment.user.nickname}</span>
                        <span>${utils.formatDate(comment.created_at)}</span>
                    </div>
                    ${comment.user.id === currentUserId ? `
                        <div class="flex">
                            <button class="edit-comment-btn">Edit</button>
                            <button class="delete-comment-btn">Delete</button>
                        </div>
                    ` : ''}
                </div>
                <div class="comment-content">
                    <p>${comment.content}</p>
                </div>
                <form class="edit-comment-form hidden" data-comment-id="${comment.id}">
                    <div class="flex">
                        <input type="text" value="${comment.content}" name="content" required>
                        <button type="submit">Save</button>
                        <button type="button" class="cancel-edit-btn">Cancel</button>
                    </div>
                </form>
            </div>
        `;
    }
};

export { uiComponents };