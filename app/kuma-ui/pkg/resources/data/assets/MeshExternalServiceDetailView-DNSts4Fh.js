import{d as X,r as s,o as i,p as l,w as e,b as n,Q as f,e as t,t as p,q as u,_ as D}from"./index-BIN9nSPF.js";import{_ as R}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3fFCInp0.js";const A=X({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(C){const o=C;return(c,r)=>{const h=s("XBadge"),b=s("XAction"),v=s("KumaPort"),x=s("XAboutCard"),z=s("DataSource"),w=s("XLayout"),E=s("AppView"),k=s("RouteView");return i(),l(k,{name:"mesh-external-service-detail-view",params:{codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:d,can:V,t:m})=>[n(E,null,{default:e(()=>[n(w,{type:"stack"},{default:e(()=>[n(x,{title:m("services.mesh-external-service.about.title"),created:o.data.creationTime,modified:o.data.modificationTime},{default:e(()=>[o.data.namespace.length>0?(i(),l(f,{key:0,layout:"horizontal"},{title:e(()=>[t(p(m("http.api.property.namespace")),1)]),body:e(()=>[n(h,{appearance:"decorative"},{default:e(()=>[t(p(o.data.namespace),1)]),_:1})]),_:2},1024)):u("",!0),r[4]||(r[4]=t()),V("use zones")&&o.data.zone?(i(),l(f,{key:1,layout:"horizontal"},{title:e(()=>[t(p(m("http.api.property.zone")),1)]),body:e(()=>[n(h,{appearance:"decorative"},{default:e(()=>[n(b,{to:{name:"zone-cp-detail-view",params:{zone:o.data.zone}}},{default:e(()=>[t(p(o.data.zone),1)]),_:1},8,["to"])]),_:1})]),_:2},1024)):u("",!0),r[5]||(r[5]=t()),c.data.spec.match?(i(),l(f,{key:2,layout:"horizontal",class:"port"},{title:e(()=>[t(p(m("http.api.property.port")),1)]),body:e(()=>[n(v,{port:c.data.spec.match},null,8,["port"])]),_:2},1024)):u("",!0),r[6]||(r[6]=t()),c.data.spec.match?(i(),l(f,{key:3,layout:"horizontal",class:"tls"},{title:e(()=>[t(p(m("http.api.property.tls")),1)]),body:e(()=>{var a;return[n(h,{appearance:(a=c.data.spec.tls)!=null&&a.enabled?"success":"neutral"},{default:e(()=>{var _;return[t(p((_=c.data.spec.tls)!=null&&_.enabled?"Enabled":"Disabled"),1)]}),_:1},8,["appearance"])]}),_:2},1024)):u("",!0)]),_:2},1032,["title","created","modified"]),r[7]||(r[7]=t()),n(R,{resource:o.data.config,"is-searchable":"",query:d.params.codeSearch,"is-filter-mode":d.params.codeFilter,"is-reg-exp-mode":d.params.codeRegExp,onQueryChange:a=>d.update({codeSearch:a}),onFilterModeChange:a=>d.update({codeFilter:a}),onRegExpModeChange:a=>d.update({codeRegExp:a})},{default:e(({copy:a,copying:_})=>[_?(i(),l(z,{key:0,src:`/meshes/${o.data.mesh}/mesh-external-service/${o.data.id}/as/kubernetes?no-store`,onChange:y=>{a(g=>g(y))},onError:y=>{a((g,S)=>S(y))}},null,8,["src","onChange","onError"])):u("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}}),B=D(A,[["__scopeId","data-v-8fae7813"]]);export{B as default};
