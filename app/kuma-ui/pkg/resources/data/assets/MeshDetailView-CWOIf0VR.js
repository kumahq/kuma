import{d as E,m as c,o as l,w as e,b as t,e as o,r as i,p as w,c as f,v as V,F as _,L as A,t as m,R as C,s as M,I as x}from"./index-D_WxlpfD.js";import{_ as I}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-jaVrNBcO.js";const q=E({__name:"MeshDetailView",props:{mesh:{}},setup(L){const a=L;return(z,s)=>{const R=i("RouteTitle"),X=i("XI18n"),g=i("XNotification"),D=i("XAction"),h=i("XBadge"),N=i("XAboutCard"),y=i("XLayout"),B=i("XCard"),b=i("DataSource"),P=i("AppView"),S=i("RouteView");return l(),c(S,{name:"mesh-detail-view",params:{mesh:""}},{default:e(({route:k,t:r,uri:v})=>[t(R,{title:r("meshes.routes.overview.title"),render:!1},null,8,["title"]),s[9]||(s[9]=o()),t(b,{src:v(M(x),"/mesh-insights/:name",{name:k.params.mesh})},{default:e(({data:n})=>[t(P,{docs:r("meshes.href.docs"),notifications:!0},{default:e(()=>[a.mesh.mtlsBackend?w("",!0):(l(),c(g,{key:0,uri:`meshes.notifications.mtls-warning:${a.mesh.id}`},{default:e(()=>[t(X,{path:"meshes.notifications.mtls-warning"})]),_:1},8,["uri"])),s[8]||(s[8]=o()),t(y,{type:"stack"},{default:e(()=>[t(N,{title:r("meshes.routes.item.about.title"),created:a.mesh.creationTime,modified:a.mesh.modificationTime},{default:e(()=>[(l(),f(_,null,V(["MeshTrafficPermission","MeshMetric","MeshAccessLog","MeshTrace"],u=>{var d;return l(),f(_,{key:u},[(l(!0),f(_,null,V([((d=n==null?void 0:n.policies)==null?void 0:d[u])??{total:0}],p=>(l(),f(_,{key:typeof p},[u==="MeshTrafficPermission"&&a.mesh.mtlsBackend&&p.total===0?(l(),c(g,{key:0,uri:`meshes.notifications.mtp-warning:${a.mesh.id}`},{default:e(()=>[t(X,{path:"meshes.notifications.mtp-warning"})]),_:1},8,["uri"])):w("",!0),s[1]||(s[1]=o()),t(A,{layout:"horizontal"},{title:e(()=>[t(D,{to:{name:"policy-list-view",params:{mesh:k.params.mesh,policyPath:`${u.toLowerCase()}s`}}},{default:e(()=>[o(m(u),1)]),_:2},1032,["to"])]),body:e(()=>[t(h,{appearance:p.total>0?"success":"neutral"},{default:e(()=>[o(m(p.total>0?r("meshes.detail.enabled"):r("meshes.detail.disabled")),1)]),_:2},1032,["appearance"])]),_:2},1024)],64))),128))],64)}),64)),s[3]||(s[3]=o()),t(A,{layout:"horizontal"},{title:e(()=>[o(m(r("http.api.property.mtls")),1)]),body:e(()=>[a.mesh.mtlsBackend?(l(),c(h,{key:1,appearance:"info"},{default:e(()=>[o(m(a.mesh.mtlsBackend.type)+" / "+m(a.mesh.mtlsBackend.name),1)]),_:1})):(l(),c(h,{key:0,appearance:"neutral"},{default:e(()=>[o(m(r("meshes.detail.disabled")),1)]),_:2},1024))]),_:2},1024)]),_:2},1032,["title","created","modified"]),s[6]||(s[6]=o()),t(B,null,{default:e(()=>[t(y,{type:"stack"},{default:e(()=>[t(y,{type:"columns",class:"columns-with-borders"},{default:e(()=>[t(C,{total:(n==null?void 0:n.services.total)??0,"data-testid":"services-status"},{title:e(()=>[o(m(r("meshes.detail.services")),1)]),_:2},1032,["total"]),s[4]||(s[4]=o()),t(C,{total:(n==null?void 0:n.dataplanesByType.standard.total)??0,online:(n==null?void 0:n.dataplanesByType.standard.online)??0,"data-testid":"data-plane-proxies-status"},{title:e(()=>[o(m(r("meshes.detail.data_plane_proxies")),1)]),_:2},1032,["total","online"]),s[5]||(s[5]=o()),t(C,{total:(n==null?void 0:n.totalPolicyCount)??0,"data-testid":"policies-status"},{title:e(()=>[o(m(r("meshes.detail.policies")),1)]),_:2},1032,["total"])]),_:2},1024)]),_:2},1024)]),_:2},1024),s[7]||(s[7]=o()),t(B,null,{default:e(()=>[t(I,{resource:a.mesh.config},{default:e(({copy:u,copying:d})=>[d?(l(),c(b,{key:0,src:v(M(x),"/meshes/:name/as/kubernetes",{name:k.params.mesh},{cacheControl:"no-store"}),onChange:p=>{u(T=>T(p))},onError:p=>{u((T,$)=>$(p))}},null,8,["src","onChange","onError"])):w("",!0)]),_:2},1032,["resource"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["docs"])]),_:2},1032,["src"])]),_:1})}}});export{q as default};
