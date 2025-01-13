import{d as M,r as a,o as l,p,w as t,b as r,l as m,e as n,Q as h,c as v,J as y,K as C,t as u,q as X}from"./index-BIN9nSPF.js";import{_ as z}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3fFCInp0.js";const $={class:"stack"},A={class:"stack-with-borders"},L={class:"mt-4"},K=M({__name:"MeshMultiZoneServiceSummaryView",props:{items:{}},setup(x){const w=x;return(N,o)=>{const S=a("RouteTitle"),R=a("XAction"),V=a("KumaPort"),_=a("XLayout"),k=a("XBadge"),E=a("DataSource"),b=a("AppView"),B=a("DataCollection"),D=a("RouteView");return l(),p(D,{name:"mesh-multi-zone-service-summary-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:s,t:g})=>[r(B,{items:w.items,predicate:i=>i.id===s.params.service},{item:t(({item:i})=>[r(b,null,{title:t(()=>[m("h2",null,[r(R,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:s.params.mesh,service:s.params.service}}},{default:t(()=>[r(S,{title:g("services.routes.item.title",{name:i.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:t(()=>[o[7]||(o[7]=n()),m("div",$,[m("div",A,[r(h,{layout:"horizontal"},{title:t(()=>o[0]||(o[0]=[n(`
                  Ports
                `)])),body:t(()=>[r(_,{type:"separated",truncate:""},{default:t(()=>[(l(!0),v(y,null,C(i.spec.ports,e=>(l(),p(V,{key:e.port,port:{...e,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),_:2},1024),o[4]||(o[4]=n()),r(h,{layout:"horizontal"},{title:t(()=>o[2]||(o[2]=[n(`
                  Selector
                `)])),body:t(()=>[r(_,{type:"separated",truncate:""},{default:t(()=>[(l(!0),v(y,null,C(i.spec.selector.meshService.matchLabels,(e,c)=>(l(),p(k,{key:`${c}:${e}`,appearance:"info"},{default:t(()=>[n(u(c)+":"+u(e),1)]),_:2},1024))),128))]),_:2},1024)]),_:2},1024)]),o[6]||(o[6]=n()),m("div",null,[m("h3",null,u(g("services.routes.item.config")),1),o[5]||(o[5]=n()),m("div",L,[r(z,{resource:i.config,"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter,"is-reg-exp-mode":s.params.codeRegExp,onQueryChange:e=>s.update({codeSearch:e}),onFilterModeChange:e=>s.update({codeFilter:e}),onRegExpModeChange:e=>s.update({codeRegExp:e})},{default:t(({copy:e,copying:c})=>[c?(l(),p(E,{key:0,src:`/meshes/${s.params.mesh}/mesh-multi-zone-service/${s.params.service}/as/kubernetes?no-store`,onChange:d=>{e(f=>f(d))},onError:d=>{e((f,F)=>F(d))}},null,8,["src","onChange","onError"])):X("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{K as default};
