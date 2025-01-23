import{d as I,r as a,o as r,q as d,w as e,b as o,e as n,p as V,E as A,c as f,K as h,L as B,Q as $,R as M,t as u,T as C,s as X,m as j}from"./index-U3igbuyl.js";import{_ as q}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-DqTPb1Kg.js";const F={key:0},K={key:1},G=I({__name:"MeshDetailView",props:{mesh:{}},setup(R){const p=R;return(x,t)=>{const D=a("RouteTitle"),b=a("XI18n"),E=a("XAction"),y=a("XBadge"),L=a("XAboutCard"),k=a("XLayout"),N=a("XCard"),v=a("DataSource"),P=a("AppView"),S=a("RouteView");return r(),d(S,{name:"mesh-detail-view",params:{mesh:""}},{default:e(({route:w,t:i,uri:T})=>[o(D,{title:i("meshes.routes.overview.title"),render:!1},null,8,["title"]),t[9]||(t[9]=n()),o(v,{src:T(V(A),"/mesh-insights/:name",{name:w.params.mesh})},{default:e(({data:s})=>[(r(!0),f(h,null,B([["MeshTrafficPermission","TrafficPermission"].reduce((_,m)=>{var l,c;return _+(((c=(l=s==null?void 0:s.policies)==null?void 0:l[m])==null?void 0:c.total)??0)},0)===0],_=>(r(),d(P,{key:_,docs:i("meshes.href.docs")},$({default:e(()=>[t[8]||(t[8]=n()),o(k,{type:"stack"},{default:e(()=>[o(L,{title:i("meshes.routes.item.about.title"),created:p.mesh.creationTime,modified:p.mesh.modificationTime},{default:e(()=>[(r(),f(h,null,B(["MeshTrafficPermission","MeshMetric","MeshAccessLog","MeshTrace"],m=>(r(),f(h,{key:m},[(r(!0),f(h,null,B([Object.entries((s==null?void 0:s.policies)??{}).find(([l])=>l===m)],l=>(r(),d(M,{key:l,layout:"horizontal"},{title:e(()=>[o(E,{to:{name:"policy-list-view",params:{mesh:w.params.mesh,policyPath:`${m.toLowerCase()}s`}}},{default:e(()=>[n(u(m),1)]),_:2},1032,["to"])]),body:e(()=>[o(y,{appearance:l?"success":"neutral"},{default:e(()=>[n(u(i(l?"meshes.detail.enabled":"meshes.detail.disabled")),1)]),_:2},1032,["appearance"])]),_:2},1024))),128))],64))),64)),t[3]||(t[3]=n()),o(M,{layout:"horizontal"},{title:e(()=>[n(u(i("http.api.property.mtls")),1)]),body:e(()=>[p.mesh.mtlsBackend?(r(),d(y,{key:1,appearance:"info"},{default:e(()=>[n(u(p.mesh.mtlsBackend.type)+" / "+u(p.mesh.mtlsBackend.name),1)]),_:1})):(r(),d(y,{key:0,appearance:"neutral"},{default:e(()=>[n(u(i("meshes.detail.disabled")),1)]),_:2},1024))]),_:2},1024)]),_:2},1032,["title","created","modified"]),t[6]||(t[6]=n()),o(N,null,{default:e(()=>[o(k,{type:"stack"},{default:e(()=>[o(k,{type:"columns",class:"columns-with-borders"},{default:e(()=>[o(C,{total:(s==null?void 0:s.services.total)??0,"data-testid":"services-status"},{title:e(()=>[n(u(i("meshes.detail.services")),1)]),_:2},1032,["total"]),t[4]||(t[4]=n()),o(C,{total:(s==null?void 0:s.dataplanesByType.standard.total)??0,online:(s==null?void 0:s.dataplanesByType.standard.online)??0,"data-testid":"data-plane-proxies-status"},{title:e(()=>[n(u(i("meshes.detail.data_plane_proxies")),1)]),_:2},1032,["total","online"]),t[5]||(t[5]=n()),o(C,{total:(s==null?void 0:s.totalPolicyCount)??0,"data-testid":"policies-status"},{title:e(()=>[n(u(i("meshes.detail.policies")),1)]),_:2},1032,["total"])]),_:2},1024)]),_:2},1024)]),_:2},1024),t[7]||(t[7]=n()),o(q,{resource:x.mesh.config},{default:e(({copy:m,copying:l})=>[l?(r(),d(v,{key:0,src:T(V(A),"/meshes/:name/as/kubernetes",{name:w.params.mesh},{cacheControl:"no-store"}),onChange:c=>{m(g=>g(c))},onError:c=>{m((g,z)=>z(c))}},null,8,["src","onChange","onError"])):X("",!0)]),_:2},1032,["resource"])]),_:2},1024)]),_:2},[!p.mesh.mtlsBackend||_?{name:"notifications",fn:e(()=>[j("ul",null,[p.mesh.mtlsBackend?X("",!0):(r(),f("li",F,[o(b,{path:"meshes.routes.item.mtls-warning"})])),t[0]||(t[0]=n()),p.mesh.mtlsBackend&&_?(r(),f("li",K,[o(b,{path:"meshes.routes.item.mtp-warning"})])):X("",!0)])]),key:"0"}:void 0]),1032,["docs"]))),128))]),_:2},1032,["src"])]),_:1})}}});export{G as default};
