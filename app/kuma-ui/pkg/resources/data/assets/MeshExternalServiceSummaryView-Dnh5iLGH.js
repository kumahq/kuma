import{_ as V}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-AmtNwljp.js";import{d as D,e as r,o as i,m as l,w as e,a as n,k as c,b as a,Y as m,t as p,p as d}from"./index-Bxa7bIor.js";import"./CodeBlock-CDtZQtk8.js";const B={class:"stack"},F={class:"stack-with-borders"},M={class:"mt-4"},$=D({__name:"MeshExternalServiceSummaryView",props:{items:{}},setup(v){const f=v;return(N,A)=>{const y=r("RouteTitle"),u=r("XAction"),C=r("KumaPort"),x=r("KBadge"),b=r("DataSource"),w=r("AppView"),k=r("DataCollection"),z=r("RouteView");return i(),l(z,{name:"mesh-external-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:t,t:h,can:E})=>[n(k,{items:f.items,predicate:o=>o.id===t.params.service},{item:e(({item:o})=>[n(w,null,{title:e(()=>[c("h2",null,[n(u,{to:{name:"mesh-external-service-detail-view",params:{mesh:t.params.mesh,service:t.params.service}}},{default:e(()=>[n(y,{title:h("services.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[a(),c("div",B,[c("div",F,[o.namespace.length>0?(i(),l(m,{key:0,layout:"horizontal"},{title:e(()=>[a(`
                  Namespace
                `)]),body:e(()=>[a(p(o.namespace),1)]),_:2},1024)):d("",!0),a(),E("use zones")&&o.zone?(i(),l(m,{key:1,layout:"horizontal"},{title:e(()=>[a(`
                  Zone
                `)]),body:e(()=>[n(u,{to:{name:"zone-cp-detail-view",params:{zone:o.zone}}},{default:e(()=>[a(p(o.zone),1)]),_:2},1032,["to"])]),_:2},1024)):d("",!0),a(),o.spec.match?(i(),l(m,{key:2,layout:"horizontal"},{title:e(()=>[a(`
                  Port
                `)]),body:e(()=>[n(C,{port:o.spec.match},null,8,["port"])]),_:2},1024)):d("",!0),a(),n(m,{layout:"horizontal"},{title:e(()=>[a(`
                  TLS
                `)]),body:e(()=>[n(x,{appearance:"neutral"},{default:e(()=>{var s;return[a(p((s=o.spec.tls)!=null&&s.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),_:2},1024)]),a(),c("div",null,[c("h3",null,p(h("services.routes.item.config")),1),a(),c("div",M,[n(V,{resource:o.config,"is-searchable":"",query:t.params.codeSearch,"is-filter-mode":t.params.codeFilter,"is-reg-exp-mode":t.params.codeRegExp,onQueryChange:s=>t.update({codeSearch:s}),onFilterModeChange:s=>t.update({codeFilter:s}),onRegExpModeChange:s=>t.update({codeRegExp:s})},{default:e(({copy:s,copying:R})=>[R?(i(),l(b,{key:0,src:`/meshes/${t.params.mesh}/mesh-service/${t.params.service}/as/kubernetes?no-store`,onChange:_=>{s(g=>g(_))},onError:_=>{s((g,S)=>S(_))}},null,8,["src","onChange","onError"])):d("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{$ as default};
