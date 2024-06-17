import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-SieuMvZn.js";import{d as T,i as n,o as r,a as l,w as t,j as c,g as i,k as s,R as p,b as u,H as v,J as y,t as d,e as _}from"./index-N0uhizir.js";import"./CodeBlock-Czkw7pCY.js";import"./toYaml-DB9FPXFY.js";const A={class:"stack"},M={class:"stack-with-borders"},K={class:"mt-4"},X=T({__name:"MeshExternalServiceSummaryView",props:{items:{}},setup(C){const x=C;return(N,z)=>{const k=n("RouteTitle"),b=n("XAction"),w=n("KTruncate"),h=n("KBadge"),E=n("DataSource"),R=n("AppView"),S=n("DataCollection"),V=n("RouteView");return r(),l(V,{name:"mesh-external-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:a,t:g})=>[c(S,{items:x.items,predicate:o=>o.id===a.params.service},{item:t(({item:o})=>[c(R,null,{title:t(()=>[i("h2",null,[c(b,{to:{name:"mesh-external-service-detail-view",params:{mesh:a.params.mesh,service:a.params.service}}},{default:t(()=>[c(k,{title:g("services.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[s(),i("div",A,[i("div",M,[o.status.addresses.length>0?(r(),l(p,{key:0,layout:"horizontal"},{title:t(()=>[s(`
                  Addresses
                `)]),body:t(()=>[c(w,null,{default:t(()=>[(r(!0),u(v,null,y(o.status.addresses,e=>(r(),u("span",{key:e.hostname},d(e.hostname),1))),128))]),_:2},1024)]),_:2},1024)):_("",!0),s(),o.spec.match?(r(),l(p,{key:1,layout:"horizontal"},{title:t(()=>[s(`
                  Port
                `)]),body:t(()=>[(r(!0),u(v,null,y([o.spec.match],e=>(r(),l(h,{key:e.port,appearance:"info"},{default:t(()=>[s(d(e.port)+"/"+d(e.protocol),1)]),_:2},1024))),128))]),_:2},1024)):_("",!0),s(),c(p,{layout:"horizontal"},{title:t(()=>[s(`
                  TLS
                `)]),body:t(()=>[c(h,{appearance:"neutral"},{default:t(()=>{var e;return[s(d((e=o.spec.tls)!=null&&e.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),_:2},1024)]),s(),i("div",null,[i("h3",null,d(g("services.routes.item.config")),1),s(),i("div",K,[c(F,{resource:o.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:e=>a.update({codeSearch:e}),onFilterModeChange:e=>a.update({codeFilter:e}),onRegExpModeChange:e=>a.update({codeRegExp:e})},{default:t(({copy:e,copying:D})=>[D?(r(),l(E,{key:0,src:`/meshes/${a.params.mesh}/mesh-service/${a.params.service}/as/kubernetes?no-store`,onChange:m=>{e(f=>f(m))},onError:m=>{e((f,B)=>B(m))}},null,8,["src","onChange","onError"])):_("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{X as default};
