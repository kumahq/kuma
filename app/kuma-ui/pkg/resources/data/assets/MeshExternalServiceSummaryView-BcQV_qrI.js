import{_ as T}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-B4RYDpTy.js";import{d as B,r,o as n,m as c,w as e,b as l,k as d,e as s,U as p,t as i,p as m,T as A,c as _,F as v,G as x}from"./index-D7Wwvihu.js";import"./CodeBlock-CQ0ZNdrm.js";const M={class:"stack"},K={class:"stack-with-borders"},N={class:"mt-4"},G=B({__name:"MeshExternalServiceSummaryView",props:{items:{}},setup(k){const C=k;return($,q)=>{const b=r("RouteTitle"),h=r("XAction"),w=r("KTruncate"),f=r("KBadge"),z=r("DataSource"),E=r("AppView"),R=r("DataCollection"),S=r("RouteView");return n(),c(S,{name:"mesh-external-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:o,t:g,can:V})=>[l(R,{items:C.items,predicate:t=>t.id===o.params.service},{item:e(({item:t})=>[l(E,null,{title:e(()=>[d("h2",null,[l(h,{to:{name:"mesh-external-service-detail-view",params:{mesh:o.params.mesh,service:o.params.service}}},{default:e(()=>[l(b,{title:g("services.routes.item.title",{name:t.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[s(),d("div",M,[d("div",K,[V("use zones")&&t.zone?(n(),c(p,{key:0,layout:"horizontal"},{title:e(()=>[s(`
                  Zone
                `)]),body:e(()=>[l(h,{to:{name:"zone-cp-detail-view",params:{zone:t.zone}}},{default:e(()=>[s(i(t.zone),1)]),_:2},1032,["to"])]),_:2},1024)):m("",!0),s(),t.status.addresses.length>0?(n(),c(p,{key:1,layout:"horizontal"},{title:e(()=>[s(`
                  Addresses
                `)]),body:e(()=>[t.status.addresses.length===1?(n(),c(A,{key:0,text:t.status.addresses[0].hostname},{default:e(()=>[s(i(t.status.addresses[0].hostname),1)]),_:2},1032,["text"])):(n(),c(w,{key:1},{default:e(()=>[(n(!0),_(v,null,x(t.status.addresses,a=>(n(),_("span",{key:a.hostname},i(a.hostname),1))),128))]),_:2},1024))]),_:2},1024)):m("",!0),s(),t.spec.match?(n(),c(p,{key:2,layout:"horizontal"},{title:e(()=>[s(`
                  Port
                `)]),body:e(()=>[(n(!0),_(v,null,x([t.spec.match],a=>(n(),c(f,{key:a.port,appearance:"info"},{default:e(()=>[s(i(a.port)+"/"+i(a.protocol),1)]),_:2},1024))),128))]),_:2},1024)):m("",!0),s(),l(p,{layout:"horizontal"},{title:e(()=>[s(`
                  TLS
                `)]),body:e(()=>[l(f,{appearance:"neutral"},{default:e(()=>{var a;return[s(i((a=t.spec.tls)!=null&&a.enabled?"Enabled":"Disabled"),1)]}),_:2},1024)]),_:2},1024)]),s(),d("div",null,[d("h3",null,i(g("services.routes.item.config")),1),s(),d("div",N,[l(T,{resource:t.config,"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:a=>o.update({codeSearch:a}),onFilterModeChange:a=>o.update({codeFilter:a}),onRegExpModeChange:a=>o.update({codeRegExp:a})},{default:e(({copy:a,copying:D})=>[D?(n(),c(z,{key:0,src:`/meshes/${o.params.mesh}/mesh-service/${o.params.service}/as/kubernetes?no-store`,onChange:u=>{a(y=>y(u))},onError:u=>{a((y,F)=>F(u))}},null,8,["src","onChange","onError"])):m("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{G as default};
