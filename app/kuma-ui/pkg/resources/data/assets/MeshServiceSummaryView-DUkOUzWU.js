import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BTrb7rgc.js";import{d as $,r as l,o as r,m as p,w as t,b as s,k as i,e as o,T as h,c as m,F as g,s as f,t as c,p as k}from"./index--1DEc0sn.js";import"./CodeBlock-CZTqRzUy.js";import"./toYaml-DB9FPXFY.js";const B={class:"stack"},A={class:"stack-with-borders"},M={class:"mt-4"},X=$({__name:"MeshServiceSummaryView",props:{items:{}},setup(x){const w=x;return(P,K)=>{const R=l("RouteTitle"),V=l("XAction"),u=l("KTruncate"),v=l("KBadge"),E=l("DataSource"),S=l("AppView"),T=l("DataCollection"),b=l("RouteView");return r(),p(b,{name:"mesh-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:a,t:y})=>[s(T,{items:w.items,predicate:n=>n.id===a.params.service},{item:t(({item:n})=>[s(S,null,{title:t(()=>[i("h2",null,[s(V,{to:{name:"mesh-service-detail-view",params:{mesh:a.params.mesh,service:a.params.service}}},{default:t(()=>[s(R,{title:y("services.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[o(),i("div",B,[i("div",A,[n.status.addresses.length>0?(r(),p(h,{key:0,layout:"horizontal"},{title:t(()=>[o(`
                  Addresses
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),m(g,null,f(n.status.addresses,e=>(r(),m("span",{key:e.hostname},c(e.hostname),1))),128))]),_:2},1024)]),_:2},1024)):k("",!0),o(),s(h,{layout:"horizontal"},{title:t(()=>[o(`
                  Ports
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),m(g,null,f(n.spec.ports,e=>(r(),p(v,{key:e.port,appearance:"info"},{default:t(()=>[o(c(e.port)+c(e.targetPort?`:${e.targetPort}`:"")+c(e.appProtocol?`/${e.appProtocol}`:""),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024),o(),s(h,{layout:"horizontal"},{title:t(()=>[o(`
                  Dataplane Tags
                `)]),body:t(()=>[s(u,null,{default:t(()=>[(r(!0),m(g,null,f(n.spec.selector.dataplaneTags,(e,d)=>(r(),p(v,{key:`${d}:${e}`,appearance:"info"},{default:t(()=>[o(c(d)+":"+c(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),o(),i("div",null,[i("h3",null,c(y("services.routes.item.config")),1),o(),i("div",M,[s(F,{resource:n.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:e=>a.update({codeSearch:e}),onFilterModeChange:e=>a.update({codeFilter:e}),onRegExpModeChange:e=>a.update({codeRegExp:e})},{default:t(({copy:e,copying:d})=>[d?(r(),p(E,{key:0,src:`/meshes/${a.params.mesh}/mesh-service/${a.params.service}/as/kubernetes?no-store`,onChange:_=>{e(C=>C(_))},onError:_=>{e((C,D)=>D(_))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{X as default};
