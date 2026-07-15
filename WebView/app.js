// ==========================================
// 1. Client OS Detection & Download Mapper
// ==========================================

document.addEventListener("DOMContentLoaded", () => {
    detectAndSetupDownload();
});

function detectAndSetupDownload() {
    const userAgent = window.navigator.userAgent;
    const platform = window.navigator.platform;
    const osTextEl = document.getElementById("detected-os");
    const downloadBtn = document.getElementById("primary-download-btn");
    
    let detectedOS = "Windows"; // Fallback default
    let downloadPath = "https://raw.githubusercontent.com/BRO-CODES-HERE/OpenChat/main/OpenChat-Win.exe";
    let iconClass = "fa-brands fa-windows";

    if (platform.indexOf("Win") !== -1) {
        detectedOS = "Windows (.exe)";
        downloadPath = "https://raw.githubusercontent.com/BRO-CODES-HERE/OpenChat/main/OpenChat-Win.exe";
        iconClass = "fa-brands fa-windows";
    } else if (platform.indexOf("Mac") !== -1 || userAgent.indexOf("Macintosh") !== -1) {
        detectedOS = "macOS (Apple Silicon / Intel)";
        downloadPath = "https://raw.githubusercontent.com/BRO-CODES-HERE/OpenChat/main/OpenChat-Mac.zip";
        iconClass = "fa-brands fa-apple";
    } else if (platform.indexOf("Linux") !== -1 || userAgent.indexOf("Linux") !== -1) {
        detectedOS = "Linux (Binary)";
        downloadPath = "https://raw.githubusercontent.com/BRO-CODES-HERE/OpenChat/main/OpenChat-linux";
        iconClass = "fa-brands fa-linux";
    }

    // Update UI elements
    if (osTextEl && downloadBtn) {
        osTextEl.textContent = detectedOS;
        downloadBtn.href = downloadPath;
        downloadBtn.setAttribute("download", "");
        
        // Update button icon dynamically
        const iconEl = downloadBtn.querySelector(".btn-icon-wrapper i");
        if (iconEl) {
            iconEl.className = iconClass;
        }
    }

    // Check if running locally on file:// protocol
    if (window.location.protocol === 'file:') {
        const warningEl = document.getElementById("file-protocol-warning");
        if (warningEl) {
            warningEl.style.display = "flex";
        }
    }
}

// ==========================================
// 2. Interactive Code Copy Handler
// ==========================================

function copyCode(button) {
    const codeEl = button.parentElement.querySelector("code");
    if (!codeEl) return;
    
    const textToCopy = codeEl.textContent;
    navigator.clipboard.writeText(textToCopy).then(() => {
        // Switch copy icon to success checkmark
        const icon = button.querySelector("i");
        icon.className = "fa-solid fa-check";
        icon.style.color = "#4ade80"; // Bright Green
        
        setTimeout(() => {
            icon.className = "fa-regular fa-copy";
            icon.style.color = "";
        }, 2000);
    });
}

// ==========================================
// 3. Three.js P2P Network Mesh Animation
// ==========================================

const canvas = document.getElementById("p2p-canvas");
if (canvas) {
    initThreeJS();
}

function initThreeJS() {
    let width = window.innerWidth;
    let height = window.innerHeight;

    // 1. Scene setup
    const scene = new THREE.Scene();

    // 2. Camera setup
    const camera = new THREE.PerspectiveCamera(60, width / height, 1, 1000);
    camera.position.z = 200;

    // 3. Renderer setup
    const renderer = new THREE.WebGLRenderer({
        canvas: canvas,
        antialias: true,
        alpha: true
    });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));

    // 4. Create Nodes (Particles)
    const particleCount = Math.min(Math.floor(width / 15), 100); // Scale nodes based on width
    const particleGeometry = new THREE.BufferGeometry();
    const positions = new Float32Array(particleCount * 3);
    const velocities = [];

    const areaSize = 350; // Coordinates range

    for (let i = 0; i < particleCount; i++) {
        // Random placement in box
        positions[i * 3] = (Math.random() - 0.5) * areaSize;
        positions[i * 3 + 1] = (Math.random() - 0.5) * areaSize;
        positions[i * 3 + 2] = (Math.random() - 0.5) * areaSize;

        // Random velocities
        velocities.push({
            x: (Math.random() - 0.5) * 0.4,
            y: (Math.random() - 0.5) * 0.4,
            z: (Math.random() - 0.5) * 0.4
        });
    }

    particleGeometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));

    // Material for node points
    const loader = new THREE.TextureLoader();
    // Generate a clean dot texture programmatically to avoid placeholder dependencies
    const canvasDot = document.createElement('canvas');
    canvasDot.width = 16;
    canvasDot.height = 16;
    const ctx = canvasDot.getContext('2d');
    const grad = ctx.createRadialGradient(8, 8, 0, 8, 8, 8);
    grad.addColorStop(0, 'rgba(0, 243, 255, 1)');
    grad.addColorStop(0.5, 'rgba(0, 243, 255, 0.4)');
    grad.addColorStop(1, 'rgba(0, 243, 255, 0)');
    ctx.fillStyle = grad;
    ctx.fillRect(0, 0, 16, 16);
    
    const dotTexture = new THREE.CanvasTexture(canvasDot);

    const particleMaterial = new THREE.PointsMaterial({
        color: 0x00f3ff,
        size: 5,
        transparent: true,
        blending: THREE.AdditiveBlending,
        map: dotTexture,
        depthWrite: false
    });

    const pointCloud = new THREE.Points(particleGeometry, particleMaterial);
    scene.add(pointCloud);

    // 5. Lines Material & Setup
    const lineMaterial = new THREE.LineBasicMaterial({
        color: 0x7f00ff,
        transparent: true,
        opacity: 0.15,
        blending: THREE.AdditiveBlending
    });

    let lineMesh = new THREE.LineSegments(new THREE.BufferGeometry(), lineMaterial);
    scene.add(lineMesh);

    // 6. Interactive Mouse Parallax
    let mouseX = 0;
    let mouseY = 0;
    let targetX = 0;
    let targetY = 0;

    window.addEventListener('mousemove', (event) => {
        mouseX = (event.clientX - width / 2) * 0.05;
        mouseY = (event.clientY - height / 2) * 0.05;
    });

    // 7. Animation Loop
    const clock = new THREE.Clock();

    function animate() {
        requestAnimationFrame(animate);

        // Smooth camera drift response to mouse movement
        targetX += (mouseX - targetX) * 0.05;
        targetY += (mouseY - targetY) * 0.05;
        camera.position.x = targetX;
        camera.position.y = -targetY;
        camera.lookAt(scene.position);

        const positionAttr = pointCloud.geometry.attributes.position;
        const currentPositions = positionAttr.array;

        // Move particles & bounce off boundaries
        for (let i = 0; i < particleCount; i++) {
            currentPositions[i * 3] += velocities[i].x;
            currentPositions[i * 3 + 1] += velocities[i].y;
            currentPositions[i * 3 + 2] += velocities[i].z;

            // Boundary collision checks (bounce back)
            const limit = areaSize / 2;
            if (Math.abs(currentPositions[i * 3]) > limit) velocities[i].x *= -1;
            if (Math.abs(currentPositions[i * 3 + 1]) > limit) velocities[i].y *= -1;
            if (Math.abs(currentPositions[i * 3 + 2]) > limit) velocities[i].z *= -1;
        }

        positionAttr.needsUpdate = true;

        // Build dynamic connections mesh (lines between close nodes)
        const linePositions = [];
        const lineColors = [];
        const maxDist = 75; // Connect nodes closer than 75 units

        for (let i = 0; i < particleCount; i++) {
            const x1 = currentPositions[i * 3];
            const y1 = currentPositions[i * 3 + 1];
            const z1 = currentPositions[i * 3 + 2];

            for (let j = i + 1; j < particleCount; j++) {
                const x2 = currentPositions[j * 3];
                const y2 = currentPositions[j * 3 + 1];
                const z2 = currentPositions[j * 3 + 2];

                const dx = x1 - x2;
                const dy = y1 - y2;
                const dz = z1 - z2;
                const dist = Math.sqrt(dx * dx + dy * dy + dz * dz);

                if (dist < maxDist) {
                    // Node 1
                    linePositions.push(x1, y1, z1);
                    // Node 2
                    linePositions.push(x2, y2, z2);
                }
            }
        }

        // Dispose old lines geometry and create a new one to prevent GPU leaks
        scene.remove(lineMesh);
        lineMesh.geometry.dispose();

        const lineGeometry = new THREE.BufferGeometry();
        lineGeometry.setAttribute('position', new THREE.Float32BufferAttribute(linePositions, 3));
        
        lineMesh = new THREE.LineSegments(lineGeometry, lineMaterial);
        scene.add(lineMesh);

        // Rotate scene slightly over time
        pointCloud.rotation.y += 0.0005;
        lineMaterial.opacity = 0.12 + Math.sin(clock.getElapsedTime()) * 0.05; // Subtle fade pulses

        renderer.render(scene, camera);
    }

    animate();

    // 8. Responsive Resize Handler
    window.addEventListener('resize', () => {
        width = window.innerWidth;
        height = window.innerHeight;

        camera.aspect = width / height;
        camera.updateProjectionMatrix();

        renderer.setSize(width, height);
    });
}
