// web/static/js/login.js
document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('login-form');
    const errorMessage = document.getElementById('error-message');
    
    // Comprobar si ya hay un token en localStorage
    const token = localStorage.getItem('token');
    if (token) {
        // Verificar si el token es válido
        fetch('/api/users/me', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(response => {
            if (response.ok) {
                // Token válido, redirigir según el rol
                return response.json();
            } else {
                // Token inválido, eliminar
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                throw new Error('Token inválido');
            }
        })
        .then(data => {
            redirectBasedOnRole(data.role);
        })
        .catch(error => {
            console.error('Error verificando token:', error);
        });
    }
    
    loginForm.addEventListener('submit', function(e) {
        e.preventDefault();
        
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        // Limpiar mensaje de error anterior
        errorMessage.textContent = '';
        
        // Enviar solicitud de inicio de sesión
        fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                username: username,
                password: password
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Usuario o contraseña incorrectos');
            }
            return response.json();
        })
        .then(data => {
            // Guardar token y datos del usuario
            localStorage.setItem('token', data.token);
            localStorage.setItem('user', JSON.stringify(data.user));
            
            // Redirigir según el rol
            redirectBasedOnRole(data.user.role);
        })
        .catch(error => {
            errorMessage.textContent = error.message;
        });
    });
    
    // Función para redirigir según el rol
    function redirectBasedOnRole(role) {
        if (role === 'admin') {
            window.location.href = '/admin';
        } else if (role === 'teacher') {
            window.location.href = '/dashboard';
        } else {
            window.location.href = '/evaluacion';
        }
    }
});