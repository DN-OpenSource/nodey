package generator

import (
	"encoding/json"
	"fmt"
	"os"
)

// GenerateHTML writes the flowchart JSON into a template HTML file and saves the raw JSON.
func GenerateHTML(flowchartData interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(flowchartData, "", "  ")
	if err != nil {
		return err
	}

	// Save JSON file
	jsonFilename := filename + ".json"
	if len(filename) > 5 && filename[len(filename)-5:] == ".html" {
		jsonFilename = filename[:len(filename)-5] + ".json"
	}
	if err := os.WriteFile(jsonFilename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save JSON: %w", err)
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nodey</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@500;600;700&display=swap" rel="stylesheet">
    <script src="https://cdnjs.cloudflare.com/ajax/libs/dagre/0.8.5/dagre.min.js"></script>
    <style>
        :root {
            --bg-color: #F8FAFC;
            --grid-line: #E2E8F0;
            --text-main: #1E293B;
            --white: #FFFFFF;
            
            /* Node Colors */
            --start-border: #F472B6; 
            --start-shadow: rgba(244, 114, 182, 0.25);
            
            --action-border: #3B82F6;
            --action-shadow: rgba(59, 130, 246, 0.2);
            
            --decision-border: #F59E0B;
            --decision-shadow: rgba(245, 158, 11, 0.2);
            
            --end-border: #64748B;
            --end-shadow: rgba(100, 116, 139, 0.2);
        }

        body {
            margin: 0;
            font-family: 'Inter', sans-serif;
            background-color: var(--bg-color);
            background-image: 
                linear-gradient(var(--grid-line) 1px, transparent 1px),
                linear-gradient(90deg, var(--grid-line) 1px, transparent 1px);
            background-size: 50px 50px; /* Larger grid for cleaner look */
            overflow: hidden;
            width: 100vw;
            height: 100vh;
            color: var(--text-main);
            user-select: none;
        }

        /* Canvas */
        #viewport {
            width: 100%%;
            height: 100%%;
            cursor: grab;
            position: relative;
            overflow: hidden;
        }
        #viewport:active { cursor: grabbing; }

        #canvas {
            position: absolute;
            top: 0; left: 0;
            transform-origin: 0 0;
            width: 0; height: 0;
        }

        /* Nodes */
        .node {
            position: absolute;
            background: var(--white);
            display: flex;
            align-items: center;
            justify-content: center;
            text-align: center;
            font-size: 14px;
            font-weight: 600;
            color: #0F172A;
            box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1), 0 4px 6px -2px rgba(0,0,0,0.05); /* Deeper shadow */
            transition: transform 0.2s ease, box-shadow 0.2s ease;
            cursor: pointer;
            z-index: 10;
        }
        .node:hover {
            transform: scale(1.05);
            z-index: 20;
            box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1), 0 10px 10px -5px rgba(0,0,0,0.04);
        }

        /* Type: Start (Pill) */
        .node.type-start {
            width: 140px;
            height: 60px;
            border-radius: 9999px;
            border: 2px solid var(--start-border);
            background: #FDF2F8; /* Subtle tint */
        }

        /* Type: Action (Rounded Rect) */
        .node.type-action {
            width: 180px;
            height: 80px;
            border-radius: 12px;
            border: 2px solid var(--action-border);
            padding: 0 10px;
        }

        /* Type: Decision (Diamond Wrapper) */
        .node.type-decision-wrap {
            width: 140px; 
            height: 140px;
            background: none;
            box-shadow: none;
            border: none;
        }
        .diamond-shape {
            position: absolute;
            width: 100px; /* Sqrt(140^2/2) approx */
            height: 100px;
            background: var(--white);
            border: 2px solid var(--decision-border);
            transform: rotate(45deg);
            z-index: 0;
            box-shadow: 0 10px 15px -3px rgba(245, 158, 11, 0.15);
        }
        .decision-text {
            position: relative;
            z-index: 1;
            max-width: 120px;
            font-size: 13px;
        }

        /* Type: End (Circle) */
        .node.type-end {
            width: 70px; 
            height: 70px;
            border-radius: 50%%;
            border: 3px double var(--end-border);
            background: #F8FAFC;
        }

        /* SVG Connections */
        svg {
            position: absolute; 
            top: 0; left: 0;
            width: 100%%; height: 100%%;
            pointer-events: none;
            overflow: visible;
            z-index: 1;
        }
        path {
            fill: none;
            stroke: #94A3B8;
            stroke-width: 2px;
            stroke-linecap: round;
            stroke-linejoin: round;
        }
        
        /* Labels on lines */
        .edge-label-bg {
            fill: var(--bg-color);
            opacity: 0.9;
        }
        .edge-label {
            font-size: 11px;
            font-weight: 700;
            fill: #64748B;
            text-anchor: middle;
        }

        /* Sidebar Inspector */
        #inspector {
            position: fixed;
            top: 24px; right: 24px;
            width: 340px;
            max-height: 85vh;
            background: rgba(255, 255, 255, 0.8);
            backdrop-filter: blur(12px);
            border-radius: 16px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
            border: 1px solid rgba(255,255,255,0.5);
            transform: translateX(400px);
            transition: transform 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            z-index: 100;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        #inspector.visible { transform: translateX(0); }
        
        .ins-header {
            padding: 24px;
            border-bottom: 1px solid rgba(0,0,0,0.05);
            display: flex;
            justify-content: space-between;
            align-items: center;
            background: rgba(255,255,255,0.5);
        }
        .ins-title { font-weight: 800; font-size: 18px; color: #0F172A; margin: 0; }
        .ins-close { cursor: pointer; border: none; background: none; font-size: 24px; color: #94A3B8; }
        .ins-body { padding: 24px; overflow-y: auto; font-size: 15px; line-height: 1.6; color: #475569; }
        .ins-body pre { 
            background: #1E293B; 
            color: #E2E8F0;
            padding: 16px; 
            border-radius: 8px; 
            margin-top: 16px; 
            white-space: pre-wrap; 
            font-family: 'Menlo', monospace; 
            font-size: 12px;
            border: 1px solid #334155;
        }

        /* Zoom Controls */
        #zoom-controls {
            position: fixed;
            bottom: 32px; left: 32px;
            background: var(--white);
            border-radius: 12px;
            box-shadow: 0 10px 15px -3px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
            overflow: hidden;
            z-index: 100;
            border: 1px solid #E2E8F0;
        }
        .zoom-btn {
            width: 44px; height: 44px;
            border: none; background: #fff;
            font-size: 20px; color: #475569;
            cursor: pointer;
            display: flex; align-items: center; justify-content: center;
            transition: background 0.2s;
        }
        .zoom-btn:hover { background: #F1F5F9; color: #0F172A; }
        #zoom-val {
            font-size: 12px;
            text-align: center;
            color: #64748B;
            border-top: 1px solid #F1F5F9;
            border-bottom: 1px solid #F1F5F9;
            padding: 8px 0;
            font-weight: 700;
            background: #F8FAFC;
        }

    </style>
</head>
<body>

    <div id="viewport">
        <div id="canvas">
            <svg id="lines">
                <defs>
                    <marker id="arrow" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto" markerUnits="strokeWidth">
                        <path d="M0,0 L0,6 L9,3 z" fill="#94A3B8" />
                    </marker>
                    <marker id="arrow-yes" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                         <path d="M0,0 L0,6 L9,3 z" fill="#10B981" />
                    </marker>
                    <marker id="arrow-no" markerWidth="10" markerHeight="10" refX="9" refY="3" orient="auto">
                         <path d="M0,0 L0,6 L9,3 z" fill="#EF4444" />
                    </marker>
                </defs>
            </svg>
            <div id="nodes"></div>
        </div>
    </div>

    <div id="inspector">
        <div class="ins-header">
            <h3 class="ins-title" id="ins-title">Details</h3>
            <button class="ins-close" onclick="closeInspector()">&times;</button>
        </div>
        <div class="ins-body" id="ins-body">Select a node...</div>
    </div>

    <div id="zoom-controls">
        <button class="zoom-btn" onclick="updateZoom(0.1)">+</button>
        <div id="zoom-val">100%%</div>
        <button class="zoom-btn" onclick="updateZoom(-0.1)">-</button>
    </div>

    <script>
        const data = %s;

        // Init Dagre
        const g = new dagre.graphlib.Graph();
        g.setGraph({ 
            rankdir: 'TB', 
            nodesep: 140, // Horizontal spacing
            ranksep: 120, // Vertical spacing
            edgesep: 50
        });
        g.setDefaultEdgeLabel(function() { return {}; });

        // Add Nodes
        data.nodes.forEach(node => {
            let w = 180, h = 80;
            if (node.type === 'start') { w = 140; h = 60; }
            if (node.type === 'end') { w = 70; h = 70; }
            if (node.type === 'decision') { w = 140; h = 140; }
            
            g.setNode(node.id, { label: node.title, width: w, height: h, type: node.type, notes: node.notes });
        });

        // Add Edges
        data.connections.forEach(conn => {
            g.setEdge(conn.from, conn.to, { type: conn.type });
        });

        // Layout
        dagre.layout(g);

        // Render
        const nodesDiv = document.getElementById('nodes');
        const linesSvg = document.getElementById('lines');

        g.nodes().forEach(id => {
            const n = g.node(id);
            const el = document.createElement('div');
            el.id = 'node-' + id;
            
            if (n.type === 'decision') {
                el.className = 'node type-decision-wrap';
                el.innerHTML = '<div class="diamond-shape"></div><div class="decision-text">' + n.label + '</div>';
            } else {
                el.className = 'node type-' + n.type;
                el.innerText = n.label;
            }

            // Center using top/left minus half width/height
            const x = n.x - n.width/2;
            const y = n.y - n.height/2;
            
            el.style.left = x + 'px';
            el.style.top = y + 'px';
            el.dataset.title = n.label;
            el.dataset.notes = n.notes || '';
            el.dataset.type = n.type;
            
            el.onclick = (e) => {
                e.stopPropagation();
                showInspector(el.dataset);
            };

            nodesDiv.appendChild(el);
        });

        g.edges().forEach(e => {
            const edge = g.edge(e);
            // Dagre gives points
            const points = edge.points;
            
            let d = "M " + points[0].x + " " + points[0].y;
            // Use smooth curve through points
            // Simple Catmull-Rom or Cubic Bezier logic?
            // Dagre points usually are distinct control points in polyline
            // Let's create a smooth curve using simplified logic or just polyline with rounded corners
            
            // For true smooth curves, we can use the points as L commands but rounded.
            // Or better: cubic bezier. 
            // Simple robust approach: Straight lines with rounded corners or just straight for cleanliness.
            // User requested "clean".
            
            // Let's try smooth cubic from start to end with control points from dagre? No dagre gives waypoints.
            // Let's use a mapping of points to svg command.
            
            if (points.length > 2) {
                 // Curve through middle points
                 // Using 'C' if possible? No, 'L' is safer for now, maybe rounded joins.
                 // To make it look "Great" let's try a smoothing function.
                 
                 // Render as simple polyline for clean "Engineering" look, or basic curve.
                 // Let's try a simple Bezier roughly following the path.
                 // Actually, standard orthogonal/curved library lines look best.
                 // We will simply draw lines to intermediate points.
                 
                 for(let i=1; i<points.length; i++) {
                     d += " L " + points[i].x + " " + points[i].y;
                 }
            } else {
                 d += " L " + points[1].x + " " + points[1].y;
            }

            const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
            path.setAttribute("d", d);
            path.setAttribute("stroke", "#94A3B8");
            
            // Marker / Color
            let c = edge.type;
            if (c === 'yes') {
                path.setAttribute("stroke", "#10B981");
                path.setAttribute("marker-end", "url(#arrow-yes)");
            } else if (c === 'no') {
                path.setAttribute("stroke", "#EF4444");
                path.setAttribute("marker-end", "url(#arrow-no)");
            } else {
                path.setAttribute("marker-end", "url(#arrow)");
            }
            
            // Add slight curve radius to CSS 'stroke-linejoin: round' takes care of corners visually
            linesSvg.appendChild(path);

            // Labels for decision
            if (c === 'yes' || c === 'no') {
                // Determine label pos - usually near first 25% of path or points[0]
                const lp = points[0]; // Start
                const np = points[1] || points[0];
                
                const txt = document.createElementNS("http://www.w3.org/2000/svg", "text");
                // Offset slightly
                txt.setAttribute("x", (lp.x + np.x)/2 + 10);
                txt.setAttribute("y", (lp.y + np.y)/2);
                txt.textContent = c.toUpperCase();
                txt.setAttribute("class", "edge-label");
                if(c==='yes') txt.setAttribute("fill", "#10B981");
                else txt.setAttribute("fill", "#EF4444");
                
                // Background for text (optional, simpler not to collide)
                linesSvg.appendChild(txt);
            }
        });

        // Center View
        const gW = g.graph().width;
        const gH = g.graph().height;
        let scale = 1;
        let panX = (window.innerWidth - gW)/2;
        let panY = (window.innerHeight - gH)/2;
        panY = Math.max(50, panY); // Ensure top margin

        const canvas = document.getElementById('canvas');
        
        function updateTransform() {
            canvas.style.transform = 'translate(' + panX + 'px, ' + panY + 'px) scale(' + scale + ')';
        }
        updateTransform();

        // Pan Zoom
        let isPanning = false;
        let startX, startY;
        
        document.getElementById('viewport').onmousedown = function(e) {
            isPanning = true;
            startX = e.clientX - panX;
            startY = e.clientY - panY;
            document.body.style.cursor = 'grabbing';
        };
        
        window.onmousemove = function(e) {
            if(!isPanning) return;
            e.preventDefault();
            panX = e.clientX - startX;
            panY = e.clientY - startY;
            updateTransform();
        };
        
        window.onmouseup = function() {
            isPanning = false;
            document.body.style.cursor = 'default';
        };
        
        window.onwheel = function(e) {
            e.preventDefault();
            scale += e.deltaY * -0.001;
            scale = Math.min(Math.max(0.1, scale), 4);
            document.getElementById('zoom-val').innerText = Math.round(scale*100) + '%%';
            updateTransform();
        };

        window.updateZoom = function(delta) {
             scale = Math.min(Math.max(0.1, scale + delta), 4);
             document.getElementById('zoom-val').innerText = Math.round(scale*100) + '%%';
             updateTransform();
        };

        // Inspector
        const inspector = document.getElementById('inspector');
        function showInspector(ds) {
             document.getElementById('ins-title').innerText = ds.title;
             let html = '<p><strong>Type:</strong> ' + ds.type + '</p>';
             html += '<p><strong>Logic:</strong></p><pre>' + (ds.notes || 'No details') + '</pre>';
             document.getElementById('ins-body').innerHTML = html;
             inspector.classList.add('visible');
        }
        window.closeInspector = function() {
             inspector.classList.remove('visible');
        }

    </script>
</body>
</html>`, string(jsonData))

	return os.WriteFile(filename, []byte(htmlContent), 0644)
}
