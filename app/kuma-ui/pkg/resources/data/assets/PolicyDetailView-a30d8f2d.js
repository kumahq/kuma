import{d as $,g as b,y as B,r as u,o as e,l as n,j as i,k as h,ay as L,n as _,w as o,i as s,E as v,x,af as P,F as S,I as C,H as R,s as N,R as A,p as q,$ as E,a5 as I}from"./index-adcc6fc8.js";import{_ as j}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-1849c988.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-4a09da2d.js";import"./toYaml-4e00099e.js";const F={key:3,"data-testid":"affected-data-plane-proxies"},H=$({__name:"PolicyConnections",props:{mesh:{type:String,required:!0},policyPath:{type:String,required:!0},policyName:{type:String,required:!0}},setup(y){const t=y,{t:f}=b(),r=B("");return(k,p)=>{const c=u("RouterLink");return e(),n("div",null,[i(h(L),{id:"dataplane-search",modelValue:r.value,"onUpdate:modelValue":p[0]||(p[0]=a=>r.value=a),type:"text",placeholder:h(f)("policies.detail.dataplane_input_placeholder"),required:"","data-testid":"dataplane-search-input"},null,8,["modelValue","placeholder"]),_(),i(N,{src:`/meshes/${t.mesh}/policy-path/${t.policyPath}/policy/${t.policyName}/dataplanes`},{default:o(({data:a,error:m})=>[m?(e(),s(v,{key:0,error:m},null,8,["error"])):a===void 0?(e(),s(x,{key:1})):a.items.length===0?(e(),s(P,{key:2})):(e(),n("ul",F,[(e(!0),n(S,null,C(a.items.filter(l=>l.dataplane.name.toLowerCase().includes(r.value.toLowerCase())),(l,w)=>(e(),n("li",{key:w,"data-testid":"dataplane-name"},[i(c,{to:{name:"data-plane-detail-view",params:{mesh:l.dataplane.mesh,dataPlane:l.dataplane.name}}},{default:o(()=>[_(R(l.dataplane.name),1)]),_:2},1032,["to"])]))),128))]))]),_:1},8,["src"])])}}}),K={class:"stack"},O=$({__name:"PolicyDetails",props:{policy:{type:Object,required:!0},path:{type:String,required:!0}},setup(y){const t=y,{t:f}=b(),r=A();return(k,p)=>(e(),n("div",K,[i(h(E),null,{body:o(()=>[q("h2",null,R(h(f)("policies.detail.affected_dpps")),1),_(),i(H,{class:"mt-4",mesh:t.policy.mesh,"policy-name":t.policy.name,"policy-path":t.path},null,8,["mesh","policy-name","policy-path"])]),_:1}),_(),i(j,{id:"code-block-policy",resource:t.policy,"resource-fetcher":c=>h(r).getSinglePolicyEntity({name:t.policy.name,mesh:t.policy.mesh,path:t.path},c),"is-searchable":""},null,8,["resource","resource-fetcher"])]))}}),G=$({__name:"PolicyDetailView",setup(y){return(t,f)=>{const r=u("RouteTitle"),k=u("DataSource"),p=u("AppView"),c=u("RouteView");return e(),s(c,{name:"policy-detail-view",params:{mesh:"",policy:"",policyPath:""}},{default:o(({route:a,t:m})=>[i(k,{src:"/*/policy-types"},{default:o(({data:l,error:w})=>[w?(e(),s(v,{key:0,error:w},null,8,["error"])):l===void 0?(e(),s(x,{key:1})):l.policies.length===0?(e(),s(P,{key:2})):(e(!0),n(S,{key:3},C([l.policies.find(d=>d.path===a.params.policyPath)??l.policies[0]],d=>(e(),s(p,{key:d.name,breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:a.params.mesh}},text:a.params.mesh},{to:{name:"policy-list-view",params:{mesh:a.params.mesh,policyPath:a.params.policyPath}},text:m("policies.routes.item.breadcrumbs")}]},{title:o(()=>[q("h1",null,[i(I,{text:a.params.policy},{default:o(()=>[i(r,{title:m("policies.routes.item.title",{name:a.params.policy}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:o(()=>[_(),i(k,{src:`/meshes/${a.params.mesh}/policy-path/${d.path}/policy/${a.params.policy}`},{default:o(({data:g,error:V})=>[V?(e(),s(v,{key:0,error:V},null,8,["error"])):g===void 0?(e(),s(x,{key:1})):(e(),s(O,{key:2,policy:g,path:d.path,"data-testid":"detail-view-details"},null,8,["policy","path"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"]))),128))]),_:2},1024)]),_:1})}}});export{G as default};
