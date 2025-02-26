import{d as L,r as a,o as n,m as u,w as t,b as o,U as z,e as s,t as d,c as _,F as h,v,p as N,aq as T,C as $,s as q,q as P,_ as K}from"./index-C-Llvxgw.js";import{_ as Q}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-D7DuEGVs.js";const Z=L({__name:"MeshMultiZoneServiceDetailView",props:{data:{}},setup(w){const p=w;return(g,r)=>{const x=a("KumaPort"),C=a("XBadge"),X=a("XAboutCard"),D=a("XCopyButton"),S=a("XAction"),b=a("XLayout"),V=a("DataCollection"),R=a("DataLoader"),k=a("XCard"),B=a("DataSource"),E=a("AppView"),M=a("RouteView");return n(),u(M,{name:"mesh-multi-zone-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:i,t:c,uri:A,me:y})=>[o(E,null,{default:t(()=>[o(b,{type:"stack"},{default:t(()=>[o(X,{title:c("services.mesh-multi-zone-service.about.title"),created:p.data.creationTime,modified:p.data.modificationTime},{default:t(()=>[o(z,{layout:"horizontal"},{title:t(()=>[s(d(c("http.api.property.ports")),1)]),body:t(()=>[p.data.spec.ports.length?(n(!0),_(h,{key:0},v(p.data.spec.ports,e=>(n(),u(x,{key:e.port,port:{...e,targetPort:void 0}},null,8,["port"]))),128)):(n(),_(h,{key:1},[s(d(c("common.detail.none")),1)],64))]),_:2},1024),r[2]||(r[2]=s()),o(z,{layout:"horizontal"},{title:t(()=>[s(d(c("http.api.property.selector")),1)]),body:t(()=>[Object.keys(g.data.spec.selector.meshService.matchLabels).length?(n(!0),_(h,{key:0},v(g.data.spec.selector.meshService.matchLabels,(e,l)=>(n(),u(C,{key:`${l}:${e}`,appearance:"info"},{default:t(()=>[s(d(l)+":"+d(e),1)]),_:2},1024))),128)):(n(),_(h,{key:1},[s(d(c("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1032,["title","created","modified"]),r[5]||(r[5]=s()),o(k,null,{title:t(()=>[s(d(c("services.detail.hostnames.title")),1)]),default:t(()=>[r[4]||(r[4]=s()),o(R,{src:A(N(T),"/meshes/:mesh/:serviceType/:serviceName/_hostnames",{mesh:i.params.mesh,serviceType:"meshmultizoneservices",serviceName:i.params.service})},{loadable:t(({data:e})=>[o(V,{type:"hostnames",items:(e==null?void 0:e.items)??[void 0]},{default:t(()=>[o($,{type:"hostnames-collection","data-testid":"hostnames-collection",items:e==null?void 0:e.items,headers:[{...y.get("headers.hostname"),label:c("services.detail.hostnames.hostname"),key:"hostname"},{...y.get("headers.zones"),label:c("services.detail.hostnames.zone"),key:"zones"}],onResize:y.set},{hostname:t(({row:l})=>[q("b",null,[o(D,{text:l.hostname},null,8,["text"])])]),zones:t(({row:l})=>[o(b,{type:"separated"},{default:t(()=>[(n(!0),_(h,null,v(l.zones,(m,f)=>(n(),u(C,{key:f,appearance:"decorative"},{default:t(()=>[o(S,{to:{name:"zone-cp-detail-view",params:{zone:m.name}}},{default:t(()=>[s(d(m.name),1)]),_:2},1032,["to"])]),_:2},1024))),128))]),_:2},1024)]),_:2},1032,["items","headers","onResize"])]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024),r[6]||(r[6]=s()),o(k,null,{default:t(()=>[o(Q,{resource:p.data.config,"is-searchable":"",query:i.params.codeSearch,"is-filter-mode":i.params.codeFilter,"is-reg-exp-mode":i.params.codeRegExp,onQueryChange:e=>i.update({codeSearch:e}),onFilterModeChange:e=>i.update({codeFilter:e}),onRegExpModeChange:e=>i.update({codeRegExp:e})},{default:t(({copy:e,copying:l})=>[l?(n(),u(B,{key:0,src:`/meshes/${p.data.mesh}/mesh-multi-zone-service/${p.data.id}/as/kubernetes?no-store`,onChange:m=>{e(f=>f(m))},onError:m=>{e((f,F)=>F(m))}},null,8,["src","onChange","onError"])):P("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:1})}}}),O=K(Z,[["__scopeId","data-v-de858d6e"]]);export{O as default};
