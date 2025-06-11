import authService from './authService.js';
class PostService {
    async createPost(formData) {
        const response = await fetch('/api/posts/create', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: formData
        });
        if (!response.ok) {
            throw new Error('Failed to create post');
        }
        return await response.json();
    }

    async getPosts() {
        const response = await fetch('/api/posts', {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });
        if (!response.ok) {
            throw new Error('Failed to load posts');
        }
        const data = await response.json();
        return data.posts || [];
    }

   async getCategories() {
    const response = await fetch('/api/categories', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authService.getToken()}`
        }
    });
    if (!response.ok) {
        throw new Error('Failed to load categories');
    }
    const data = await response.json();
    return data.categories;
}

    async addComment(postId, content) {
    const response = await fetch(`/api/posts/comments/${postId}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${authService.getToken()}`
        },
        body: JSON.stringify({ content })
    });
    if (!response.ok) {
        let errorMsg = 'Failed to add comment';
        try {
            const error = await response.json();
            errorMsg = error.message || errorMsg;
        } catch {
            try {
                errorMsg = await response.text();
            } catch {}
        }
        throw new Error(errorMsg);
    }
    return await response.json();
}

    async updateComment(commentId, content) {
        const response = await fetch(`/api/comments/${commentId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: JSON.stringify({ content })
        });
        if (!response.ok) {
            throw new Error('Failed to update comment');
        }
        return await response.json();
    }

    async deleteComment(commentId) {
        const response = await fetch(`/api/comments/${commentId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${authService.getToken()}`
            }
        });
        if (!response.ok) {
            throw new Error('Failed to delete comment');
        }
    }

    async handleReaction(postId, reactionType) {
        const response = await fetch(`/api/posts/react/${postId}`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${authService.getToken()}`
            },
            body: JSON.stringify({ type: reactionType })
        });
        if (!response.ok) {
            throw new Error('Failed to update reaction');
        }
        return await response.json();
    }
}
const postService = new PostService();
export default postService;