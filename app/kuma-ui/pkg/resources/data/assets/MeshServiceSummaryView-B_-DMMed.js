import{_ as D}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BkeCm4yq.js";import{d as F,r as l,o,m as i,w as t,b as n,k as d,e as r,U as h,T as B,t as c,c as m,F as g,G as f,p as k}from"./index-Is4zmHdk.js";import"./CodeBlock-DvLuvw_5.js";const A={class:"stack"},M={class:"stack-with-borders"},P={class:"mt-4"},X=F({__name:"MeshServiceSummaryView",props:{items:{}},setup(x){const w=x;return(K,N)=>{const R=l("RouteTitle"),V=l("XAction"),u=l("KTruncate"),v=l("KBadge"),E=l("DataSource"),S=l("AppView"),T=l("DataCollection"),$=l("RouteView");return o(),i($,{name:"mesh-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:a,t:y})=>[n(T,{items:w.items,predicate:s=>s.id===a.params.service},{item:t(({item:s})=>[n(S,null,{title:t(()=>[d("h2",null,[n(V,{to:{name:"mesh-service-detail-view",params:{mesh:a.params.mesh,service:a.params.service}}},{default:t(()=>[n(R,{title:y("services.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[r(),d("div",A,[d("div",M,[s.status.addresses.length>0?(o(),i(h,{key:0,layout:"horizontal"},{title:t(()=>[r(`
                  Addresses
                `)]),body:t(()=>[s.status.addresses.length===1?(o(),i(B,{key:0,text:s.status.addresses[0].hostname},{default:t(()=>[r(c(s.status.addresses[0].hostname),1)]),_:2},1032,["text"])):(o(),i(u,{key:1},{default:t(()=>[(o(!0),m(g,null,f(s.status.addresses,e=>(o(),m("span",{key:e.hostname},c(e.hostname),1))),128))]),_:2},1024))]),_:2},1024)):k("",!0),r(),n(h,{layout:"horizontal"},{title:t(()=>[r(`
                  Ports
                `)]),body:t(()=>[n(u,null,{default:t(()=>[(o(!0),m(g,null,f(s.spec.ports,e=>(o(),i(v,{key:e.port,appearance:"info"},{default:t(()=>[r(c(e.port)+c(e.targetPort?`:${e.targetPort}`:"")+c(e.appProtocol?`/${e.appProtocol}`:""),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024),r(),n(h,{layout:"horizontal"},{title:t(()=>[r(`
                  Dataplane Tags
                `)]),body:t(()=>[n(u,null,{default:t(()=>[(o(!0),m(g,null,f(s.spec.selector.dataplaneTags,(e,p)=>(o(),i(v,{key:`${p}:${e}`,appearance:"info"},{default:t(()=>[r(c(p)+":"+c(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),r(),d("div",null,[d("h3",null,c(y("services.routes.item.config")),1),r(),d("div",P,[n(D,{resource:s.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:e=>a.update({codeSearch:e}),onFilterModeChange:e=>a.update({codeFilter:e}),onRegExpModeChange:e=>a.update({codeRegExp:e})},{default:t(({copy:e,copying:p})=>[p?(o(),i(E,{key:0,src:`/meshes/${a.params.mesh}/mesh-service/${a.params.service}/as/kubernetes?no-store`,onChange:_=>{e(C=>C(_))},onError:_=>{e((C,b)=>b(_))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{X as default};
