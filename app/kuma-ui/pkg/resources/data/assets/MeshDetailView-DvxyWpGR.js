import{d as S,r as m,o as i,p as d,w as e,b as n,e as o,m as X,D as g,c as f,J as h,K as B,R as z,Q as V,t as u,T as C,q as T,l as $}from"./index-gI7YoWPY.js";import{_ as j}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CfFLyldn.js";const q=["innerHTML"],F=["innerHTML"],O=S({__name:"MeshDetailView",props:{mesh:{}},setup(A){const c=A;return(L,t)=>{const D=m("RouteTitle"),R=m("XAction"),y=m("XBadge"),x=m("XAboutCard"),k=m("XLayout"),H=m("XCard"),M=m("DataSource"),N=m("AppView"),E=m("RouteView");return i(),d(E,{name:"mesh-detail-view",params:{mesh:""}},{default:e(({route:w,t:r,uri:b})=>[n(D,{title:r("meshes.routes.overview.title"),render:!1},null,8,["title"]),t[9]||(t[9]=o()),n(M,{src:b(X(g),"/mesh-insights/:name",{name:w.params.mesh})},{default:e(({data:s})=>[(i(!0),f(h,null,B([["MeshTrafficPermission","TrafficPermission"].reduce((_,a)=>{var l,p;return _+(((p=(l=s==null?void 0:s.policies)==null?void 0:l[a])==null?void 0:p.total)??0)},0)===0],_=>(i(),d(N,{key:_,docs:r("meshes.href.docs")},z({default:e(()=>[t[8]||(t[8]=o()),n(k,{type:"stack"},{default:e(()=>[n(x,{title:r("meshes.routes.item.about.title"),created:c.mesh.creationTime,modified:c.mesh.modificationTime},{default:e(()=>[(i(),f(h,null,B(["MeshTrafficPermission","MeshMetric","MeshAccessLog","MeshTrace"],a=>(i(),f(h,{key:a},[(i(!0),f(h,null,B([Object.entries((s==null?void 0:s.policies)??{}).find(([l])=>l===a)],l=>(i(),d(V,{key:l,layout:"horizontal"},{title:e(()=>[n(R,{to:{name:"policy-list-view",params:{mesh:w.params.mesh,policyPath:`${a.toLowerCase()}s`}}},{default:e(()=>[o(u(a),1)]),_:2},1032,["to"])]),body:e(()=>[n(y,{appearance:l?"success":"neutral"},{default:e(()=>[o(u(r(l?"meshes.detail.enabled":"meshes.detail.disabled")),1)]),_:2},1032,["appearance"])]),_:2},1024))),128))],64))),64)),t[3]||(t[3]=o()),n(V,{layout:"horizontal"},{title:e(()=>[o(u(r("http.api.property.mtls")),1)]),body:e(()=>[c.mesh.mtlsBackend?(i(),d(y,{key:1,appearance:"info"},{default:e(()=>[o(u(c.mesh.mtlsBackend.type)+" / "+u(c.mesh.mtlsBackend.name),1)]),_:1})):(i(),d(y,{key:0,appearance:"neutral"},{default:e(()=>[o(u(r("meshes.detail.disabled")),1)]),_:2},1024))]),_:2},1024)]),_:2},1032,["title","created","modified"]),t[6]||(t[6]=o()),n(H,null,{default:e(()=>[n(k,{type:"stack"},{default:e(()=>[n(k,{type:"columns",class:"columns-with-borders"},{default:e(()=>[n(C,{total:(s==null?void 0:s.services.total)??0,"data-testid":"services-status"},{title:e(()=>[o(u(r("meshes.detail.services")),1)]),_:2},1032,["total"]),t[4]||(t[4]=o()),n(C,{total:(s==null?void 0:s.dataplanesByType.standard.total)??0,online:(s==null?void 0:s.dataplanesByType.standard.online)??0,"data-testid":"data-plane-proxies-status"},{title:e(()=>[o(u(r("meshes.detail.data_plane_proxies")),1)]),_:2},1032,["total","online"]),t[5]||(t[5]=o()),n(C,{total:(s==null?void 0:s.totalPolicyCount)??0,"data-testid":"policies-status"},{title:e(()=>[o(u(r("meshes.detail.policies")),1)]),_:2},1032,["total"])]),_:2},1024)]),_:2},1024)]),_:2},1024),t[7]||(t[7]=o()),n(j,{resource:L.mesh.config},{default:e(({copy:a,copying:l})=>[l?(i(),d(M,{key:0,src:b(X(g),"/meshes/:name/as/kubernetes",{name:w.params.mesh},{cacheControl:"no-store"}),onChange:p=>{a(v=>v(p))},onError:p=>{a((v,P)=>P(p))}},null,8,["src","onChange","onError"])):T("",!0)]),_:2},1032,["resource"])]),_:2},1024)]),_:2},[!c.mesh.mtlsBackend||_?{name:"notifications",fn:e(()=>[$("ul",null,[c.mesh.mtlsBackend?T("",!0):(i(),f("li",{key:0,innerHTML:r("meshes.routes.item.mtls-warning")},null,8,q)),t[0]||(t[0]=o()),c.mesh.mtlsBackend&&_?(i(),f("li",{key:1,innerHTML:r("meshes.routes.item.mtp-warning")},null,8,F)):T("",!0)])]),key:"0"}:void 0]),1032,["docs"]))),128))]),_:2},1032,["src"])]),_:1})}}});export{O as default};
