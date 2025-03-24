import{d as L,r as n,m,o as c,w as e,b as t,e as o,p as h,L as v,t as l,s as M,an as N,C as T,c as q,F as $,v as K,q as P,_ as Q}from"./index-D_WxlpfD.js";import{_ as I}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-jaVrNBcO.js";const j=L({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(x){const r=x;return(_,s)=>{const y=n("XBadge"),C=n("XAction"),k=n("KumaPort"),w=n("XAboutCard"),X=n("XCopyButton"),z=n("XLayout"),D=n("DataCollection"),E=n("DataLoader"),b=n("XCard"),V=n("DataSource"),R=n("AppView"),B=n("RouteView");return c(),m(B,{name:"mesh-external-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:i,can:S,t:d,uri:A,me:g})=>[t(R,null,{default:e(()=>[t(z,{type:"stack"},{default:e(()=>[t(w,{title:d("services.mesh-external-service.about.title"),created:r.data.creationTime,modified:r.data.modificationTime},{default:e(()=>[r.data.namespace.length>0?(c(),m(v,{key:0,layout:"horizontal"},{title:e(()=>[o(l(d("http.api.property.namespace")),1)]),body:e(()=>[t(y,{appearance:"decorative"},{default:e(()=>[o(l(r.data.namespace),1)]),_:1})]),_:2},1024)):h("",!0),s[4]||(s[4]=o()),S("use zones")&&r.data.zone?(c(),m(v,{key:1,layout:"horizontal"},{title:e(()=>[o(l(d("http.api.property.zone")),1)]),body:e(()=>[t(y,{appearance:"decorative"},{default:e(()=>[t(C,{to:{name:"zone-cp-detail-view",params:{zone:r.data.zone}}},{default:e(()=>[o(l(r.data.zone),1)]),_:1},8,["to"])]),_:1})]),_:2},1024)):h("",!0),s[5]||(s[5]=o()),_.data.spec.match?(c(),m(v,{key:2,layout:"horizontal",class:"port"},{title:e(()=>[o(l(d("http.api.property.port")),1)]),body:e(()=>[t(k,{port:_.data.spec.match},null,8,["port"])]),_:2},1024)):h("",!0),s[6]||(s[6]=o()),_.data.spec.match?(c(),m(v,{key:3,layout:"horizontal",class:"tls"},{title:e(()=>[o(l(d("http.api.property.tls")),1)]),body:e(()=>{var a;return[t(y,{appearance:(a=_.data.spec.tls)!=null&&a.enabled?"success":"neutral"},{default:e(()=>{var p;return[o(l((p=_.data.spec.tls)!=null&&p.enabled?"Enabled":"Disabled"),1)]}),_:1},8,["appearance"])]}),_:2},1024)):h("",!0)]),_:2},1032,["title","created","modified"]),s[9]||(s[9]=o()),t(b,null,{title:e(()=>[o(l(d("services.detail.hostnames.title")),1)]),default:e(()=>[s[8]||(s[8]=o()),t(E,{src:A(M(N),"/meshes/:mesh/:serviceType/:serviceName/_hostnames",{mesh:i.params.mesh,serviceType:"meshexternalservices",serviceName:i.params.service})},{loadable:e(({data:a})=>[t(D,{type:"hostnames",items:(a==null?void 0:a.items)??[void 0]},{default:e(()=>[t(T,{type:"hostnames-collection","data-testid":"hostnames-collection",items:a==null?void 0:a.items,headers:[{...g.get("headers.hostname"),label:d("services.detail.hostnames.hostname"),key:"hostname"},{...g.get("headers.zones"),label:d("services.detail.hostnames.zone"),key:"zones"}],onResize:g.set},{hostname:e(({row:p})=>[P("b",null,[t(X,{text:p.hostname},null,8,["text"])])]),zones:e(({row:p})=>[t(z,{type:"separated"},{default:e(()=>[(c(!0),q($,null,K(p.zones,(u,f)=>(c(),m(y,{key:f,appearance:"decorative"},{default:e(()=>[t(C,{to:{name:"zone-cp-detail-view",params:{zone:u.name}}},{default:e(()=>[o(l(u.name),1)]),_:2},1032,["to"])]),_:2},1024))),128))]),_:2},1024)]),_:2},1032,["items","headers","onResize"])]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024),s[10]||(s[10]=o()),t(b,null,{default:e(()=>[t(I,{resource:r.data.config,"is-searchable":"",query:i.params.codeSearch,"is-filter-mode":i.params.codeFilter,"is-reg-exp-mode":i.params.codeRegExp,onQueryChange:a=>i.update({codeSearch:a}),onFilterModeChange:a=>i.update({codeFilter:a}),onRegExpModeChange:a=>i.update({codeRegExp:a})},{default:e(({copy:a,copying:p})=>[p?(c(),m(V,{key:0,src:`/meshes/${r.data.mesh}/mesh-external-service/${r.data.id}/as/kubernetes?no-store`,onChange:u=>{a(f=>f(u))},onError:u=>{a((f,F)=>F(u))}},null,8,["src","onChange","onError"])):h("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:2},1024)]),_:1})}}}),J=Q(j,[["__scopeId","data-v-2a3ac1a3"]]);export{J as default};
