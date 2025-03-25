// Actualiza estas constantes con la URL de tu API
const API_BASE_URL = "/api";
const GRADES_ENDPOINT = `${API_BASE_URL}/grades`;
const STUDENTS_ENDPOINT = `${API_BASE_URL}/students`;
const SAVE_ENDPOINT = `${API_BASE_URL}/evaluation`;

// Elementos del DOM
const saveButton = document.getElementById('saveButton');
const notificationElement = document.getElementById('notification');

// Elementos del DOM para selección de grado y estudiante
const gradeSelect = document.getElementById('gradeSelect');
const studentSelect = document.getElementById('studentSelect');
const studentNameInput = document.getElementById('studentName');
const loadingIndicator = document.createElement('div');
loadingIndicator.className = 'loading';

// Actualizar el nombre en los resultados cuando cambia en el input
document.getElementById('studentName').addEventListener('input', function() {
    document.getElementById('studentNameResult').textContent = this.value || 'Sin nombre';
});

// Función para cargar los grados disponibles
async function loadGrades() {
    try {
        // Agregar indicador de carga
        gradeSelect.parentNode.appendChild(loadingIndicator);
        
        // Realizar la petición a nuestra API Go
        const response = await fetch(GRADES_ENDPOINT);
        
        if (!response.ok) {
            throw new Error('Error al cargar los grados');
        }
        
        const data = await response.json();
        
        // Limpiar opciones existentes
        gradeSelect.innerHTML = '<option value="">Seleccione un grado</option>';
        
        // Agregar cada grado como una opción
        if (data && data.length > 0) {
            data.forEach(grade => {
                const option = document.createElement('option');
                option.value = grade.name;
                option.textContent = grade.name;
                gradeSelect.appendChild(option);
            });
        }
    } catch (error) {
        console.error('Error al cargar grados:', error);
        showNotification('Error al cargar grados: ' + error.message, false);
    } finally {
        // Quitar indicador de carga
        if (loadingIndicator.parentNode) {
            loadingIndicator.parentNode.removeChild(loadingIndicator);
        }
    }
}

// Función para cargar estudiantes según el grado seleccionado
async function loadStudentsByGrade(grade) {
        try {
        // Reiniciar selector de estudiantes
        studentSelect.innerHTML = '<option value="">Seleccionando estudiantes...</option>';
        studentSelect.disabled = true;
        studentSelect.parentNode.appendChild(loadingIndicator);
        
        // Realizar la petición a nuestra API Go
        const response = await fetch(`${STUDENTS_ENDPOINT}?grade=${encodeURIComponent(grade)}`);
        
        if (!response.ok) {
            throw new Error('Error al cargar los estudiantes');
        }
        
        const data = await response.json();
        
        // Limpiar selector y añadir nueva opción por defecto
        studentSelect.innerHTML = '<option value="">Seleccione un estudiante</option>';
        
        // Agregar estudiantes como opciones
        if (data.students && data.students.length > 0) {
            data.students.forEach(student => {
                const option = document.createElement('option');
                option.value = student.name;
                option.textContent = student.name;
                studentSelect.appendChild(option);
            });
            studentSelect.disabled = false;
        } else {
            studentSelect.innerHTML = '<option value="">No hay estudiantes para este grado</option>';
        }
    } catch (error) {
        console.error('Error al cargar estudiantes:', error);
        showNotification('Error al cargar estudiantes: ' + error.message, false);
        studentSelect.innerHTML = '<option value="">Error al cargar estudiantes</option>';
    } finally {
        // Quitar indicador de carga
        if (loadingIndicator.parentNode) {
            loadingIndicator.parentNode.removeChild(loadingIndicator);
        }
    }
}

// Event listener para cambio de grado
gradeSelect.addEventListener('change', function() {
    const selectedGrade = this.value;
    
    // Limpiar estudiante seleccionado
    studentNameInput.value = '';
    document.getElementById('studentNameResult').textContent = 'Sin nombre';
    
    if (selectedGrade) {
        loadStudentsByGrade(selectedGrade);
    } else {
        studentSelect.innerHTML = '<option value="">Primero seleccione un grado</option>';
        studentSelect.disabled = true;
    }
});

// Event listener para cambio de estudiante
studentSelect.addEventListener('change', function() {
    const selectedStudent = this.value;
    studentNameInput.value = selectedStudent;
    document.getElementById('studentNameResult').textContent = selectedStudent || 'Sin nombre';
});


// Función para mostrar notificaciones
function showNotification(message, isSuccess) {
    notificationElement.textContent = message;
    notificationElement.className = isSuccess ? 'notification success' : 'notification error';
    notificationElement.style.display = 'block';
    
    // Ocultar después de 5 segundos
    setTimeout(() => {
        notificationElement.style.display = 'none';
    }, 5000);
}

// Función para actualizar la barra de progreso
function updateProgressBar(average) {
    const progressBar = document.getElementById('averageProgressBar');
    const progressTooltip = document.getElementById('progressTooltip');
    const percentage = (average / 5) * 100;
    progressBar.style.width = percentage + '%';
    progressBar.textContent = average.toFixed(2);
    
    // Cambiar color y texto según el valor
    if (average < 3) {
        progressBar.style.backgroundColor = '#e74c3c'; // Rojo
        progressTooltip.textContent = 'Bajo';
    } else if (average < 4) {
        progressBar.style.backgroundColor = '#f39c12'; // Naranja
        progressTooltip.textContent = 'Básico';
    } else if (average < 4.6) {
        progressBar.style.backgroundColor = '#2ecc71'; // Verde
        progressTooltip.textContent = 'Alto';
    } else {
        progressBar.style.backgroundColor = '#27ae60'; // Verde oscuro
        progressTooltip.textContent = 'Superior';
    }
}

saveButton.addEventListener('click', async function() {
    const studentName = document.getElementById('studentName').value;
    if (!studentName.trim()) {
        showNotification('Por favor ingrese el nombre del estudiante', false);
        return;
    }

    // Obtener el grado seleccionado
    const selectedGrade = document.getElementById('gradeSelect').value;
    if (!selectedGrade) {
        showNotification('Por favor seleccione un grado', false);
        return;
    }
    
    // Recopilar calificaciones para enviar
    let ratings = {};
    let allRatings = [];
    
    const conceptNames = [
        "Participación activa en clase",
        "Respeto a compañeros y profesores",
        "Puntualidad en la entrega de tareas",
        "Trabajo en equipo",
        "Organización y disciplina",
        "Asistencia regular a clases",
        "Actitud positiva hacia el aprendizaje",
        "Capacidad para seguir instrucciones",
        "Comportamiento durante actividades grupales",
        "Compromiso con su propio desarrollo académico"
    ];
    
    let sum = 0;
    let ratedCount = 0;
    
    for (let i = 1; i <= 10; i++) {
        const ratingName = 'rating' + i;
        const selectedRating = document.querySelector(`input[name="${ratingName}"]:checked`);
        
        if (!selectedRating) {
            showNotification('Por favor complete todas las calificaciones', false);
            return;
        }
        
        const value = parseInt(selectedRating.value);
        
        ratings[ratingName] = value;
        allRatings.push(value);
        sum += value;
        ratedCount++;
        
        // Agregar también con los nombres de los conceptos
        //ratings[`concept${i}`] = {
        //    name: conceptNames[i-1],
        //    rating: value
        //};
    }
    
    const average = ratedCount > 0 ? sum / ratedCount : 0;
    
    // Crear objeto de datos a enviar
    const dataToSend = {
        studentName: studentName,
        grade: selectedGrade,
        ratings: ratings,
        average: Math.round(average * 100) / 100,
        date: new Date().toISOString()
    };
    
    // Deshabilitar botón durante el envío
    saveButton.disabled = true;
    saveButton.textContent = "Guardando...";
    
    try {
        const response = await fetch(SAVE_ENDPOINT, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(dataToSend)
        });
        
        if (!response.ok) {
            throw new Error('Error al guardar la evaluación');
        }
        
        const result = await response.json();
        showNotification(result.message || 'Evaluación guardada correctamente', true);
        
        // Opcional: Resetear el formulario después de guardar
        resetForm();
        
    } catch (error) {
        console.error('Error:', error);
        showNotification('Error al guardar: ' + error.message, false);
    } finally {
        saveButton.disabled = false;
        saveButton.textContent = "Guardar Calificaciones";
    }
});

function resetForm() {
    // Limpiar las calificaciones
    const ratingInputs = document.querySelectorAll('.rating-input:checked');
    ratingInputs.forEach(input => {
        input.checked = false;
    });
    
    // Actualizar resultados
    updateResults();
}

// Función para calcular y actualizar resultados
function updateResults() {
    // Recopilar todos los valores
    let ratings = [];
    let sum = 0;
    let ratedCount = 0;
    
    for (let i = 1; i <= 10; i++) {
        const ratingName = 'rating' + i;
        const selectedRating = document.querySelector(`input[name="${ratingName}"]:checked`);
        
        if (selectedRating) {
            const value = parseInt(selectedRating.value);
            ratings[i-1] = value; // Guardar en posición correspondiente
            sum += value;
            ratedCount++;
        } else {
            ratings[i-1] = "-"; // Marcar como no calificado
        }
    }
    
    // Calcular promedio
    const average = ratedCount > 0 ? sum / ratedCount : 0;
    
    // Actualizar conteo de calificados
    document.getElementById('ratedCount').textContent = ratedCount;
    
    // Activar/desactivar botón de guardar
    saveButton.disabled = ratedCount < 10;
    
    // Mostrar resultados
    document.getElementById('averageResult').textContent = average.toFixed(2);
    document.getElementById('individualRatings').textContent = ratings.join(', ');
    
    // Actualizar barra de progreso
    updateProgressBar(average);
    
    // Añadir efecto de iluminación
    const resultContainer = document.getElementById('resultContainer');
    resultContainer.classList.remove('highlight');
    void resultContainer.offsetWidth; // Truco para reiniciar la animación
    resultContainer.classList.add('highlight');
}

// Añadir event listeners a todos los inputs de calificación
const ratingInputs = document.querySelectorAll('.rating-input');
ratingInputs.forEach(input => {
    input.addEventListener('change', updateResults);
});

// Cargar grados al iniciar la página
document.addEventListener('DOMContentLoaded', function() {
    loadGrades();
    
    // El resto de la inicialización existente
    updateResults();
});