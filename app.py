import streamlit as st
import pydgraph
import json
import networkx as nx
from pyvis.network import Network
import streamlit.components.v1 as components
import tempfile
import os
import glob
import hashlib
from code_editor import code_editor

# ---------------------------------------------------------
# 1. CONFIGURATION & CONNECTION
# ---------------------------------------------------------
st.set_page_config(page_title="Dgraph Explorer", layout="wide")

@st.cache_resource
def get_dgraph_client():
    # Update this if your Dgraph is running elsewhere
    client_stub = pydgraph.DgraphClientStub("localhost:9080")
    return pydgraph.DgraphClient(client_stub)

client = get_dgraph_client()

# ---------------------------------------------------------
# 2. HELPER FUNCTIONS
# ---------------------------------------------------------
def run_query(query, vars=None):
    """Executes the query against Dgraph."""
    try:
        txn = client.txn(read_only=True)
        res = txn.query(query, variables=vars or {})
        return json.loads(res.json)
    except Exception as e:
        st.error(f"Query Error: {e}")
        return None

def get_color_from_string(string_val):
    """Generates a consistent hex color from a string."""
    if not string_val:
        return "#97c2fc" 
    hash_object = hashlib.md5(string_val.encode())
    hex_hash = hash_object.hexdigest()
    return f"#{hex_hash[:6]}"

def create_details_html(data_dict, title="Details"):
    """Formats a dictionary into a clean HTML block for the sidebar panel."""
    rows = []
    for k, v in data_dict.items():
        if isinstance(v, (list, dict)): continue
        rows.append(f"""
            <div style='margin-bottom: 10px; border-bottom: 1px solid #eee; padding-bottom: 5px;'>
                <span style='font-weight:bold; display:block; color:#666; font-size: 0.85em; text-transform: uppercase;'>{k}</span>
                <span style='color:#222; word-break: break-word; font-family: monospace;'>{v}</span>
            </div>
        """)
    
    return f"""
    <div>
        <h3 style='margin-top:0; border-bottom:2px solid #ff5733; padding-bottom:10px; margin-bottom: 15px; color:#333;'>{title}</h3>
        <div>{ "".join(rows) }</div>
    </div>
    """

def generate_pyvis_html(dgraph_res):
    """Converts Dgraph JSON to a PyVis HTML string with Click-to-Show Details."""
    G = nx.DiGraph()
    
    data = dgraph_res.get('q', [])
    if not data:
        return None

    for parent in data:
        p_uid = parent.get('uid')
        p_label = parent.get('label', p_uid)
        
        # 1. Parent Node
        p_attrs = {k: v for k, v in parent.items() if k != 'rel'}
        p_details = create_details_html(p_attrs, title=f"Node: {p_label}")
        
        # NOTE: We store the details in 'details_html' instead of 'title' to disable hover
        G.add_node(p_uid, label=p_label, details_html=p_details, color="#ff5733", size=30)
        
        relationships = parent.get('rel', [])
        for child in relationships:
            c_uid = child.get('uid')
            c_label = child.get('label', c_uid)
            
            # 2. Separate Edge Data from Node Data
            edge_attrs = {k: v for k, v in child.items() if k.startswith('rel|')}
            node_attrs = {k: v for k, v in child.items() if not k.startswith('rel|')}
            
            edge_label_text = edge_attrs.get('rel|label', '')
            visual_color = get_color_from_string(edge_label_text)
            
            # 3. Target Node
            c_details = create_details_html(node_attrs, title=f"Node: {c_label}")
            G.add_node(c_uid, label=c_label, details_html=c_details, color=visual_color, size=20)
            
            # 4. Edge
            e_details = create_details_html(edge_attrs, title=f"Edge: {edge_label_text}")
            G.add_edge(
                p_uid, 
                c_uid, 
                label=edge_label_text,
                details_html=e_details, # Custom attribute for our JS
                color=visual_color,
                font={'align': 'middle', 'size': 10}
            )

    # Configure PyVis
    net = Network(height="600px", width="100%", bgcolor="#222222", font_color="white")
    net.from_nx(G)
    net.force_atlas_2based(gravity=-50, central_gravity=0.01, spring_length=100, spring_strength=0.08, damping=0.4, overlap=0)

    # Generate base HTML
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=".html") as tmp:
            net.save_graph(tmp.name)
            path = tmp.name
        
        with open(path, 'r', encoding='utf-8') as f:
            html_string = f.read()
        
        os.remove(path)

        # ---------------------------------------------------------
        # INJECT CUSTOM CSS & JS FOR SIDEBAR PANEL
        # ---------------------------------------------------------
        
        # CSS for the 'Tab on the left'
        custom_css = """
        <style>
            #info-panel {
                position: absolute;
                top: 10px;
                left: 10px;
                bottom: 10px;
                width: 300px;
                background: rgba(255, 255, 255, 0.95);
                border-radius: 8px;
                box-shadow: 0 0 15px rgba(0,0,0,0.3);
                padding: 20px;
                overflow-y: auto;
                font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
                font-size: 14px;
                color: #333;
                z-index: 1000;
                display: none; /* Hidden by default */
                transition: all 0.3s ease;
            }
            #close-panel {
                position: absolute;
                top: 10px;
                right: 15px;
                cursor: pointer;
                font-size: 20px;
                font-weight: bold;
                color: #888;
            }
            #close-panel:hover { color: #333; }
        </style>
        <div id="info-panel">
            <div id="close-panel" onclick="document.getElementById('info-panel').style.display='none'">&times;</div>
            <div id="panel-content">Click a node or edge to see details here.</div>
        </div>
        """

        # JS to handle clicks
        custom_js = """
        <script type="text/javascript">
            // Wait for network to be initialized
            network.on("click", function (params) {
                var panel = document.getElementById('info-panel');
                var content = document.getElementById('panel-content');
                var didSelect = false;

                // 1. Check Nodes
                if (params.nodes.length > 0) {
                    var nodeId = params.nodes[0];
                    var nodeData = nodes.get(nodeId);
                    if (nodeData && nodeData.details_html) {
                        content.innerHTML = nodeData.details_html;
                        panel.style.display = 'block';
                        didSelect = true;
                    }
                } 
                
                // 2. Check Edges (if no node selected)
                if (!didSelect && params.edges.length > 0) {
                    var edgeId = params.edges[0];
                    var edgeData = edges.get(edgeId);
                    if (edgeData && edgeData.details_html) {
                        content.innerHTML = edgeData.details_html;
                        panel.style.display = 'block';
                        didSelect = true;
                    }
                }

                // 3. If background clicked, hide panel
                if (!didSelect) {
                    panel.style.display = 'none';
                }
            });
        </script>
        """
        
        # Insert CSS before body end
        html_string = html_string.replace("</body>", f"{custom_css}{custom_js}</body>")
        
        return html_string

    except Exception as e:
        st.error(f"Visualization Error: {e}")
        return None

def load_gql_files(directory="gql"):
    """Loads .gql and .graphql files from a directory into a dictionary."""
    queries = {}
    if not os.path.exists(directory):
        os.makedirs(directory)
        st.sidebar.warning(f"Created missing directory: '{directory}'. Please add .gql files there.")
        return {"Example (No files found)": "query { q(func: uid(0x123)) { uid } }"}

    files = glob.glob(os.path.join(directory, "*.gql")) + glob.glob(os.path.join(directory, "*.graphql"))
    if not files:
        return {"No .gql files found": ""}

    for filepath in files:
        filename = os.path.basename(filepath)
        with open(filepath, "r") as f:
            queries[filename] = f.read()
    return queries

# ---------------------------------------------------------
# 3. SIDEBAR: CONTROLS
# ---------------------------------------------------------
st.sidebar.header("Query Configuration")

query_map = load_gql_files("gql")
selected_filename = st.sidebar.selectbox("Select Query File", list(query_map.keys()))
target_uid = st.sidebar.text_input("Target UID ($id)", value="0x123")

# ---------------------------------------------------------
# 4. MAIN AREA
# ---------------------------------------------------------
st.title("🕸️ Dgraph Visualizer")

default_query_content = query_map.get(selected_filename, "")

custom_btns = [{
    "name": "Run",
    "feather": "Play",
    "primary": True,
    "hasText": True,
    "alwaysOn": True,
    "commands": ["submit"], 
    "style": {"bottom": "0.44rem", "right": "0.4rem"}
}]

editor_response = code_editor(
    default_query_content,
    lang="graphql",
    height=250,
    theme="monokai",
    buttons=custom_btns,
    key=f"editor_{selected_filename}",
    options={"wrap": True, "fontSize": 14, "showLineNumbers": True, "highlightActiveLine": True, "tabSize": 2}
)

query_input = editor_response['text']
should_run = editor_response['type'] == "submit"


# ---------------------------------------------------------
# 5. EXECUTION LOGIC
# ---------------------------------------------------------
if should_run:
    with st.spinner("Querying Dgraph..."):
        variables = {"$id": target_uid}
        result_data = run_query(query_input, variables)

        if result_data:
            tab_graph, tab_json = st.tabs(["🕸️ Graph Visualization", "📄 Raw JSON"])

            with tab_graph:
                html_data = generate_pyvis_html(result_data)
                if html_data:
                    components.html(html_data, height=600, scrolling=False)
                else:
                    st.warning("No graph data found in response (check if 'q' is empty).")

            with tab_json:
                st.json(result_data)