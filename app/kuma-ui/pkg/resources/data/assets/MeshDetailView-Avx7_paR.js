import{$ as E}from"./app-layout.es-BA5CIo3a.js";import{d as K,e as d,o as i,m as f,w as s,a as r,b as o,l as y,D as g,c as p,J as h,K as w,R as P,k,Q as D,t as c,T as B,p as T,q as S}from"./index-DpJ_igul.js";import{_ as j}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-DFu2BeDp.js";const q=["innerHTML"],F=["innerHTML"],I={class:"stack"},J={class:"columns"},O={class:"stack"},Q={class:"columns"},z=K({__name:"MeshDetailView",props:{mesh:{}},setup(x){const l=x;return(R,t)=>{const A=d("RouteTitle"),L=d("XAction"),C=d("XBadge"),H=d("KCard"),M=d("DataSource"),N=d("AppView"),X=d("RouteView");return i(),f(X,{name:"mesh-detail-view",params:{mesh:""}},{default:s(({route:v,t:n,uri:V})=>[r(A,{title:n("meshes.routes.overview.title"),render:!1},null,8,["title"]),t[9]||(t[9]=o()),r(M,{src:V(y(g),"/mesh-insights/:name",{name:v.params.mesh})},{default:s(({data:e})=>[(i(!0),p(h,null,w([["MeshTrafficPermission","TrafficPermission"].reduce((_,a)=>{var m,u;return _+(((u=(m=e==null?void 0:e.policies)==null?void 0:m[a])==null?void 0:u.total)??0)},0)===0],_=>(i(),f(N,{key:_,docs:n("meshes.href.docs")},P({default:s(()=>[t[8]||(t[8]=o()),k("div",I,[r(y(E),{title:n("meshes.routes.item.subtitle",{name:l.mesh.name}),created:n("common.formats.datetime",{value:Date.parse(l.mesh.creationTime)}),modified:n("common.formats.datetime",{value:Date.parse(l.mesh.modificationTime)})},{default:s(()=>[k("div",J,[(i(),p(h,null,w(["MeshTrafficPermission","MeshMetric","MeshAccessLog","MeshTrace"],a=>(i(),p(h,{key:a},[(i(!0),p(h,null,w([Object.entries((e==null?void 0:e.policies)??{}).find(([m,u])=>m===a)],m=>(i(),f(D,{key:m},{title:s(()=>[r(L,{to:{name:"policy-list-view",params:{mesh:v.params.mesh,policyPath:`${a.toLowerCase()}s`}}},{default:s(()=>[o(c(a),1)]),_:2},1032,["to"])]),body:s(()=>[r(C,{appearance:"neutral"},{default:s(()=>[o(c(n(m?"meshes.detail.enabled":"meshes.detail.disabled")),1)]),_:2},1024)]),_:2},1024))),128))],64))),64))])]),_:2},1032,["title","created","modified"]),t[6]||(t[6]=o()),r(H,null,{default:s(()=>[k("div",O,[k("div",Q,[r(B,{total:(e==null?void 0:e.services.total)??0,"data-testid":"services-status"},{title:s(()=>[o(c(n("meshes.detail.services")),1)]),_:2},1032,["total"]),t[3]||(t[3]=o()),r(B,{total:(e==null?void 0:e.dataplanesByType.standard.total)??0,online:(e==null?void 0:e.dataplanesByType.standard.online)??0,"data-testid":"data-plane-proxies-status"},{title:s(()=>[o(c(n("meshes.detail.data_plane_proxies")),1)]),_:2},1032,["total","online"]),t[4]||(t[4]=o()),r(B,{total:(e==null?void 0:e.totalPolicyCount)??0,"data-testid":"policies-status"},{title:s(()=>[o(c(n("meshes.detail.policies")),1)]),_:2},1032,["total"]),t[5]||(t[5]=o()),r(D,null,{title:s(()=>[o(c(n("http.api.property.mtls")),1)]),body:s(()=>[l.mesh.mtlsBackend?(i(),p(h,{key:1},[o(c(l.mesh.mtlsBackend.type)+" / "+c(l.mesh.mtlsBackend.name),1)],64)):(i(),f(C,{key:0,appearance:"neutral"},{default:s(()=>[o(c(n("meshes.detail.disabled")),1)]),_:2},1024))]),_:2},1024)])])]),_:2},1024),t[7]||(t[7]=o()),r(j,{resource:R.mesh.config},{default:s(({copy:a,copying:m})=>[m?(i(),f(M,{key:0,src:V(y(g),"/meshes/:name/as/kubernetes",{name:v.params.mesh},{cacheControl:"no-store"}),onChange:u=>{a(b=>b(u))},onError:u=>{a((b,$)=>$(u))}},null,8,["src","onChange","onError"])):T("",!0)]),_:2},1032,["resource"])])]),_:2},[!l.mesh.mtlsBackend||_?{name:"notifications",fn:s(()=>[k("ul",null,[l.mesh.mtlsBackend?T("",!0):(i(),p("li",{key:0,innerHTML:n("meshes.routes.item.mtls-warning")},null,8,q)),t[0]||(t[0]=o()),l.mesh.mtlsBackend&&_?(i(),p("li",{key:1,innerHTML:n("meshes.routes.item.mtp-warning")},null,8,F)):T("",!0)])]),key:"0"}:void 0]),1032,["docs"]))),128))]),_:2},1032,["src"])]),_:1})}}}),Y=S(z,[["__scopeId","data-v-069bff49"]]);export{Y as default};
