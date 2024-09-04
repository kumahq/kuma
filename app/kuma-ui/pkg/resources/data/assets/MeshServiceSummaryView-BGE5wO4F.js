import{_ as D}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-B4RYDpTy.js";import{d as F,r as c,o,m as i,w as t,b as n,k as d,e as s,U as m,t as l,p as f,T as B,c as u,F as g,G as y}from"./index-D7Wwvihu.js";import"./CodeBlock-CQ0ZNdrm.js";const A={class:"stack"},M={class:"stack-with-borders"},P={class:"mt-4"},G=F({__name:"MeshServiceSummaryView",props:{items:{}},setup(w){const z=w;return(K,N)=>{const R=c("RouteTitle"),v=c("XAction"),_=c("KTruncate"),C=c("KBadge"),V=c("DataSource"),b=c("AppView"),E=c("DataCollection"),S=c("RouteView");return o(),i(S,{name:"mesh-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:r,t:k,can:T})=>[n(E,{items:z.items,predicate:a=>a.id===r.params.service},{item:t(({item:a})=>[n(b,null,{title:t(()=>[d("h2",null,[n(v,{to:{name:"mesh-service-detail-view",params:{mesh:r.params.mesh,service:r.params.service}}},{default:t(()=>[n(R,{title:k("services.routes.item.title",{name:a.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[s(),d("div",A,[d("div",M,[T("use zones")&&a.zone?(o(),i(m,{key:0,layout:"horizontal"},{title:t(()=>[s(`
                  Zone
                `)]),body:t(()=>[n(v,{to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:t(()=>[s(l(a.zone),1)]),_:2},1032,["to"])]),_:2},1024)):f("",!0),s(),a.status.addresses.length>0?(o(),i(m,{key:1,layout:"horizontal"},{title:t(()=>[s(`
                  Addresses
                `)]),body:t(()=>[a.status.addresses.length===1?(o(),i(B,{key:0,text:a.status.addresses[0].hostname},{default:t(()=>[s(l(a.status.addresses[0].hostname),1)]),_:2},1032,["text"])):(o(),i(_,{key:1},{default:t(()=>[(o(!0),u(g,null,y(a.status.addresses,e=>(o(),u("span",{key:e.hostname},l(e.hostname),1))),128))]),_:2},1024))]),_:2},1024)):f("",!0),s(),n(m,{layout:"horizontal"},{title:t(()=>[s(`
                  Ports
                `)]),body:t(()=>[n(_,null,{default:t(()=>[(o(!0),u(g,null,y(a.spec.ports,e=>(o(),i(C,{key:e.port,appearance:"info"},{default:t(()=>[s(l(e.port)+l(e.targetPort?`:${e.targetPort}`:"")+l(e.appProtocol?`/${e.appProtocol}`:""),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024),s(),n(m,{layout:"horizontal"},{title:t(()=>[s(`
                  Dataplane Tags
                `)]),body:t(()=>[n(_,null,{default:t(()=>[(o(!0),u(g,null,y(a.spec.selector.dataplaneTags,(e,p)=>(o(),i(C,{key:`${p}:${e}`,appearance:"info"},{default:t(()=>[s(l(p)+":"+l(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),s(),d("div",null,[d("h3",null,l(k("services.routes.item.config")),1),s(),d("div",P,[n(D,{resource:a.config,"is-searchable":"",query:r.params.codeSearch,"is-filter-mode":r.params.codeFilter,"is-reg-exp-mode":r.params.codeRegExp,onQueryChange:e=>r.update({codeSearch:e}),onFilterModeChange:e=>r.update({codeFilter:e}),onRegExpModeChange:e=>r.update({codeRegExp:e})},{default:t(({copy:e,copying:p})=>[p?(o(),i(V,{key:0,src:`/meshes/${r.params.mesh}/mesh-service/${r.params.service}/as/kubernetes?no-store`,onChange:h=>{e(x=>x(h))},onError:h=>{e((x,$)=>$(h))}},null,8,["src","onChange","onError"])):f("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{G as default};
