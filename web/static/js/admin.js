// web/static/js/admin.js

document.addEventListener('DOMContentLoaded', function() {
    // Mostrar nombre de usuario
    const user = JSON.parse(localStorage.getItem('user') || '{}');
    if (user) {
        document.getElementById('username-text').textContent = user.username;
    }
    
    // Manejar cerrar sesión
    document.getElementById('logout-button').addEventListener('click', function() {
        showLoading();
        
        setTimeout(() => {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/login';
        }, 300);
    });

    // Inicializar la carga de datos
    initializeApp();
});

// Función para mostrar indicador de carga
function showLoading() {
    const loadingOverlay = document.createElement('div');
    loadingOverlay.className = 'loading-overlay';
    loadingOverlay.innerHTML = `
        <div class="loading-spinner">
            <i class="fas fa-circle-notch fa-spin"></i>
        </div>
    `;
    document.body.appendChild(loadingOverlay);
}

// Función para ocultar indicador de carga
function hideLoading() {
    const overlay = document.querySelector('.loading-overlay');
    if (overlay) {
        overlay.classList.add('fade-out');
        setTimeout(() => {
            overlay.remove();
        }, 300);
    }
}

// Función para mostrar notificaciones
function showNotification(message, type = 'success') {
    // Eliminar notificaciones existentes
    const existingNotifications = document.querySelectorAll('.notification');
    existingNotifications.forEach(notification => {
        notification.remove();
    });
    
    const notification = document.createElement('div');
    notification.className = `notification ${type}`;
    notification.innerHTML = `
        <div class="notification-content">
            <i class="fas ${type === 'success' ? 'fa-check-circle' : 'fa-exclamation-circle'}"></i>
            <span>${message}</span>
        </div>
        <button class="notification-close">&times;</button>
    `;
    document.body.appendChild(notification);
    
    // Añadir evento para cerrar la notificación
    notification.querySelector('.notification-close').addEventListener('click', () => {
        notification.remove();
    });
    
    // Eliminar notificación automáticamente después de 5 segundos
    setTimeout(() => {
        notification.classList.add('fade-out');
        setTimeout(() => {
            notification.remove();
        }, 300);
    }, 5000);
}

// Configurar event listeners para modales
function setupModalListeners() {
    // Cerrar modales al hacer clic en el botón X
    document.querySelectorAll('.modal .close').forEach(closeBtn => {
        closeBtn.addEventListener('click', function() {
            this.closest('.modal').style.display = 'none';
        });
    });
    
    // Cerrar modal al hacer clic fuera
    window.addEventListener('click', function(event) {
        document.querySelectorAll('.modal').forEach(modal => {
            if (event.target === modal) {
                modal.style.display = 'none';
            }
        });
    });
    
    // Configurar formularios
    document.getElementById('grade-form').addEventListener('submit', handleGradeSubmit);
    document.getElementById('user-form').addEventListener('submit', handleUserSubmit);
}

// Mostrar modal para agregar grado
function showAddGradeModal() {
    document.getElementById('modal-title').innerHTML = '<i class="fas fa-plus-circle"></i> Agregar Grado';
    document.getElementById('grade-id').value = '';
    document.getElementById('grade-name').value = '';
    document.getElementById('grade-active').checked = true;
    
    document.getElementById('grade-modal').style.display = 'block';
    document.getElementById('grade-name').focus();
}

// Función para manejar errores en fetch
async function handleFetchErrors(response) {
    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || response.statusText);
    }
    return response;
}

// Función para editar un grado
async function editGrade() {
    try {
        showLoading();
        const gradeId = this.dataset.id;
        const response = await fetch(`/api/admin/grades/${gradeId}`);
        await handleFetchErrors(response);
        const grade = await response.json();
        
        document.getElementById('modal-title').innerHTML = '<i class="fas fa-edit"></i> Editar Grado';
        document.getElementById('grade-id').value = grade.id;
        document.getElementById('grade-name').value = grade.name;
        document.getElementById('grade-active').checked = grade.active;
        
        document.getElementById('grade-modal').style.display = 'block';
        document.getElementById('grade-name').focus();
    } catch (error) {
        console.error('Error al obtener datos del grado:', error);
        showNotification('Error al obtener datos del grado: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Función para cambiar el estado de un grado
async function toggleGrade() {
    try {
        const gradeId = this.dataset.id;
        const isActive = this.dataset.active === 'true';
        
        if(!confirm(`¿Estás seguro de que deseas ${isActive ? 'desactivar' : 'activar'} este grado?`)) {
            return;
        }
        
        showLoading();
        
        const response = await fetch(`/api/admin/grades/${gradeId}/toggle`, {
            method: 'PUT'
        });
        await handleFetchErrors(response);
        
        showNotification(`Grado ${isActive ? 'desactivado' : 'activado'} correctamente`);
        loadGrades();
    } catch (error) {
        console.error('Error al cambiar estado del grado:', error);
        showNotification('Error al cambiar estado del grado: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Función para eliminar un grado
async function deleteGrade() {
    try {
        const gradeId = this.dataset.id;
        
        if(!confirm('¿Estás seguro de que deseas eliminar este grado? Esta acción no se puede deshacer.')) {
            return;
        }
        
        showLoading();
        
        const response = await fetch(`/api/admin/grades/${gradeId}`, {
            method: 'DELETE'
        });
        await handleFetchErrors(response);
        
        showNotification('Grado eliminado correctamente');
        loadGrades();
    } catch (error) {
        console.error('Error al eliminar grado:', error);
        showNotification('Error al eliminar grado: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Función para manejar el envío del formulario de grados
async function handleGradeSubmit(e) {
    e.preventDefault();
    
    try {
        showLoading();
        
        const gradeId = document.getElementById('grade-id').value;
        const gradeData = {
            name: document.getElementById('grade-name').value,
            active: document.getElementById('grade-active').checked
        };
        
        const url = gradeId 
            ? `/api/admin/grades/${gradeId}` 
            : '/api/admin/grades';
        
        const method = gradeId ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(gradeData)
        });
        await handleFetchErrors(response);
        
        document.getElementById('grade-modal').style.display = 'none';
        showNotification(`Grado ${gradeId ? 'actualizado' : 'creado'} correctamente`);
        loadGrades();
    } catch (error) {
        console.error('Error al guardar grado:', error);
        showNotification('Error al guardar grado: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Función para eliminar un usuario
async function deleteUser() {
    try {
        const userId = this.dataset.id;
        
        if(!confirm('¿Estás seguro de que deseas eliminar este usuario? Esta acción no se puede deshacer.')) {
            return;
        }
        
        showLoading();
        
        const response = await fetch(`/api/admin/users/${userId}`, {
            method: 'DELETE'
        });
        await handleFetchErrors(response);
        
        showNotification('Usuario eliminado correctamente');
        loadUsers();
    } catch (error) {
        console.error('Error al eliminar usuario:', error);
        showNotification('Error al eliminar usuario: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Añadir al inicio del archivo admin.js después de DOMContentLoaded

// Variables para paginación
let currentPage = 1;
let totalPages = 1;
let usersPerPage = 10;
let currentFilter = 'all';
let searchTerm = '';

// Función para filtrar usuarios
function setupUserFilters() {
    // Búsqueda de usuarios
    const searchInput = document.getElementById('user-search');
    searchInput.addEventListener('input', function() {
        searchTerm = this.value.toLowerCase();
        currentPage = 1;
        loadUsers();
    });
    
    // Filtros por rol
    document.querySelectorAll('.filter-btn').forEach(button => {
        button.addEventListener('click', function() {
            document.querySelectorAll('.filter-btn').forEach(btn => btn.classList.remove('active'));
            this.classList.add('active');
            currentFilter = this.dataset.role;
            currentPage = 1;
            loadUsers();
        });
    });
    
    // Paginación
    document.getElementById('prev-page').addEventListener('click', function() {
        if (currentPage > 1) {
            currentPage--;
            loadUsers();
        }
    });
    
    document.getElementById('next-page').addEventListener('click', function() {
        if (currentPage < totalPages) {
            currentPage++;
            loadUsers();
        }
    });
}

// Modificar la función initializeApp para incluir el setup de filtros
function initializeApp() {
    // Cargar datos iniciales
    loadGrades();
    loadUsers();
    
    // Configurar filtros de usuarios
    setupUserFilters();
    
    // Configurar event listeners para botones principales
    document.getElementById('add-grade-btn').addEventListener('click', showAddGradeModal);
    document.getElementById('add-user-btn').addEventListener('click', showAddUserModal);
    
    // Configurar event listeners para modales
    setupModalListeners();
    
    // Configurar campos condicionales
    document.getElementById('role').addEventListener('change', function() {
        const studentFields = document.getElementById('student-fields');
        const gradeSelect = document.getElementById('grade');
        
        if (this.value === 'student') {
            studentFields.classList.remove('hidden');
            gradeSelect.disabled = false;
            gradeSelect.required = true;
        } else {
            studentFields.classList.add('hidden');
            gradeSelect.disabled = true;
            gradeSelect.required = false;
        }
    });
}

// Actualizar la función loadUsers para manejar filtros y paginación
async function loadUsers() {
    try {
        const response = await fetch('/api/admin/users');
        await handleFetchErrors(response);
        let users = await response.json();
        
        // Aplicar filtros
        if (currentFilter !== 'all') {
            users = users.filter(user => user.role === currentFilter);
        }
        
        // Aplicar búsqueda
        if (searchTerm) {
            users = users.filter(user => 
                user.username.toLowerCase().includes(searchTerm) ||
                user.id.toString().includes(searchTerm)
            );
        }
        
        // Calcular paginación
        totalPages = Math.ceil(users.length / usersPerPage);
        const startIndex = (currentPage - 1) * usersPerPage;
        const paginatedUsers = users.slice(startIndex, startIndex + usersPerPage);
        
        // Actualizar controles de paginación
        document.getElementById('prev-page').disabled = currentPage <= 1;
        document.getElementById('next-page').disabled = currentPage >= totalPages;
        document.getElementById('page-info').textContent = `Página ${currentPage} de ${totalPages || 1}`;
        
        const tbody = document.querySelector('#users-table tbody');
        tbody.innerHTML = '';
        
        if (paginatedUsers.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" class="empty-message">No se encontraron usuarios${searchTerm ? ' con ese criterio de búsqueda' : ''}</td>
                </tr>
            `;
            return;
        }
        
        paginatedUsers.forEach(user => {
            // Formatear fecha (simulada ya que no está en los datos)
            const formattedDate = user.createdAt ? new Date(user.createdAt).toLocaleDateString() : 'N/A';
            
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${user.id}</td>
                <td>${user.username}</td>
                <td>
                    <span class="role-badge ${user.role}">
                        ${user.role === 'admin' ? 'Administrador' : user.role === 'teacher' ? 'Profesor' : 'Estudiante'}
                    </span>
                </td>
                <td>${formattedDate}</td>
                <td>
                    <span class="status-badge ${user.active !== false ? 'active' : 'inactive'}">
                        ${user.active !== false ? 'Activo' : 'Inactivo'}
                    </span>
                </td>
                <td class="actions-cell">
                    <button class="edit-user" data-id="${user.id}" title="Editar">
                        <i class="fas fa-edit"></i>
                    </button>
                    ${user.role === 'student' ? `
                        <button class="view-grades" data-id="${user.id}" title="Ver Calificaciones">
                            <i class="fas fa-clipboard-list"></i>
                        </button>
                    ` : ''}
                    ${user.id !== 1 ? `
                        <button class="toggle-user" data-id="${user.id}" data-active="${user.active !== false}" title="${user.active !== false ? 'Desactivar' : 'Activar'}">
                            <i class="fas ${user.active !== false ? 'fa-toggle-on' : 'fa-toggle-off'}"></i>
                        </button>
                        <button class="delete-user" data-id="${user.id}" title="Eliminar">
                            <i class="fas fa-trash-alt"></i>
                        </button>
                    ` : ''}
                </td>
            `;
            tbody.appendChild(row);
        });
        
        // Añadir event listeners para botones
        document.querySelectorAll('.edit-user').forEach(button => {
            button.addEventListener('click', editUser);
        });
        
        document.querySelectorAll('.toggle-user').forEach(button => {
            button.addEventListener('click', toggleUser);
        });
        
        document.querySelectorAll('.delete-user').forEach(button => {
            button.addEventListener('click', deleteUser);
        });
        
        document.querySelectorAll('.view-grades').forEach(button => {
            button.addEventListener('click', viewStudentGrades);
        });
    } catch (error) {
        console.error('Error al cargar los usuarios:', error);
        showNotification('Error al cargar los usuarios: ' + error.message, 'error');
    }
}

// Función para activar/desactivar usuarios
async function toggleUser() {
    try {
        const userId = this.dataset.id;
        const isActive = this.dataset.active === 'true';
        
        if(!confirm(`¿Estás seguro de que deseas ${isActive ? 'desactivar' : 'activar'} este usuario?`)) {
            return;
        }
        
        showLoading();
        
        const response = await fetch(`/api/admin/users/${userId}/toggle`, {
            method: 'PUT'
        });
        await handleFetchErrors(response);
        
        showNotification(`Usuario ${isActive ? 'desactivado' : 'activado'} correctamente`);
        loadUsers();
    } catch (error) {
        console.error('Error al cambiar estado del usuario:', error);
        showNotification('Error al cambiar estado del usuario: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Función para ver calificaciones de estudiantes
function viewStudentGrades() {
    const studentId = this.dataset.id;
    // Aquí puedes redirigir a una página de calificaciones o mostrar un modal
    // Por ejemplo:
    window.location.href = `/admin/student/${studentId}/grades`;
}

// Modificar la función showAddUserModal para incluir los nuevos campos
function showAddUserModal() {
    document.getElementById('user-modal-title').innerHTML = '<i class="fas fa-user-plus"></i> Agregar Usuario';
    document.getElementById('user-id').value = '';
    document.getElementById('username').value = '';
    document.getElementById('password').value = '';
    document.getElementById('password').required = true;
    document.getElementById('role').value = 'teacher';
    
    // Ocultar y resetear campos de estudiante
    document.getElementById('student-fields').classList.add('hidden');
    document.getElementById('grade').disabled = true;
    document.getElementById('grade').required = false;
    
    // Activar usuario por defecto
    document.getElementById('user-active').checked = true;
    
    // Mostrar/ocultar hint de contraseña
    document.querySelector('.password-hint').style.display = 'none';
    
    document.getElementById('user-modal').style.display = 'block';
    document.getElementById('username').focus();
}

// Modificar la función editUser para incluir los nuevos campos
async function editUser() {
    try {
        showLoading();
        const userId = this.dataset.id;
        const response = await fetch(`/api/admin/users/${userId}`);
        await handleFetchErrors(response);
        const user = await response.json();
        
        document.getElementById('user-modal-title').innerHTML = '<i class="fas fa-user-edit"></i> Editar Usuario';
        document.getElementById('user-id').value = user.id;
        document.getElementById('username').value = user.username;
        document.getElementById('password').value = '';
        document.getElementById('password').required = false;
        document.getElementById('role').value = user.role;
        document.getElementById('user-active').checked = user.active !== false;
        
        // Mostrar hint de contraseña
        document.querySelector('.password-hint').style.display = 'block';
        
        // Manejar campos específicos de estudiante
        if (user.role === 'student') {
            document.getElementById('student-fields').classList.remove('hidden');
            document.getElementById('grade').disabled = false;
            document.getElementById('grade').required = true;
            
            // Si el estudiante tiene un grado asignado, seleccionarlo
            if (user.gradeId) {
                document.getElementById('grade').value = user.gradeId;
            }
        } else {
            document.getElementById('student-fields').classList.add('hidden');
            document.getElementById('grade').disabled = true;
            document.getElementById('grade').required = false;
        }
        
        document.getElementById('user-modal').style.display = 'block';
        document.getElementById('username').focus();
    } catch (error) {
        console.error('Error al obtener datos del usuario:', error);
        showNotification('Error al obtener datos del usuario: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Modificar la función handleUserSubmit para incluir los nuevos campos
async function handleUserSubmit(e) {
    e.preventDefault();
    
    try {
        showLoading();
        
        const userId = document.getElementById('user-id').value;
        const userData = {
            username: document.getElementById('username').value,
            password: document.getElementById('password').value,
            role: document.getElementById('role').value,
            active: document.getElementById('user-active').checked
        };
        
        // Si el rol es estudiante, agregar el grado
        /*if (userData.role === 'student') {
            userData.gradeId = document.getElementById('grade').value;
        }*/
        
        // Si estamos editando y no se proporciona contraseña, eliminarla del objeto
        if (userId && !userData.password) {
            delete userData.password;
        }
        
        const url = userId 
            ? `/api/admin/users/${userId}` 
            : '/api/admin/users';
        
        const method = userId ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });
        await handleFetchErrors(response);
        
        document.getElementById('user-modal').style.display = 'none';
        showNotification(`Usuario ${userId ? 'actualizado' : 'creado'} correctamente`);
        loadUsers();
    } catch (error) {
        console.error('Error al guardar usuario:', error);
        showNotification('Error al guardar usuario: ' + error.message, 'error');
    } finally {
        hideLoading();
    }
}

// Modificar la función loadGrades para que actualice también el select de grados
async function loadGrades() {
    try {
        const response = await fetch('/api/admin/gradesAdmin');
        await handleFetchErrors(response);
        const grades = await response.json();
        
        // Actualizar tabla de grados (código existente)
        const tbody = document.querySelector('#grades-table tbody');
        tbody.innerHTML = '';
        
        if (grades.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="4" class="empty-message">No hay grados registrados</td>
                </tr>
            `;
        } else {
            grades.forEach(grade => {
                const row = document.createElement('tr');
                row.innerHTML = `
                    <td>${grade.id}</td>
                    <td>${grade.name}</td>
                    <td><span class="status-badge ${grade.active ? 'active' : 'inactive'}">${grade.active ? 'Activo' : 'Inactivo'}</span></td>
                    <td class="actions-cell">
                        <button class="edit-grade" data-id="${grade.id}" title="Editar">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="toggle-grade" data-id="${grade.id}" data-active="${grade.active}" title="${grade.active ? 'Desactivar' : 'Activar'}">
                            <i class="fas ${grade.active ? 'fa-toggle-on' : 'fa-toggle-off'}"></i>
                        </button>
                        <button class="delete-grade" data-id="${grade.id}" title="Eliminar">
                            <i class="fas fa-trash-alt"></i>
                        </button>
                    </td>
                `;
                tbody.appendChild(row);
            });
        }
        
        // Añadir event listeners para botones (código existente)
        document.querySelectorAll('.edit-grade').forEach(button => {
            button.addEventListener('click', editGrade);
        });
        
        document.querySelectorAll('.toggle-grade').forEach(button => {
            button.addEventListener('click', toggleGrade);
        });
        
        document.querySelectorAll('.delete-grade').forEach(button => {
            button.addEventListener('click', deleteGrade);
        });
        
        // Actualizar el select de grados para el formulario de usuarios
        const gradeSelect = document.getElementById('grade');
        gradeSelect.innerHTML = '<option value="">Seleccione un grado</option>';
        
        // Agregar solo los grados activos
        grades.filter(grade => grade.active).forEach(grade => {
            const option = document.createElement('option');
            option.value = grade.id;
            option.textContent = grade.name;
            gradeSelect.appendChild(option);
        });
    } catch (error) {
        console.error('Error al cargar los grados:', error);
        showNotification('Error al cargar los grados: ' + error.message, 'error');
    }
}