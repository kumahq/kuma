import{d as D,r as l,o as d,m as i,w as o,b as m,s as u,e as t,t as p,c as C,F as X,v as k,T as q,U as g,q as f,p as L,R as A}from"./index-Bi3CXAeE.js";import{_ as x}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-SewON5TQ.js";const N={key:0,class:"stack-with-borders","data-testid":"structured-view"},G=D({__name:"HostnameGeneratorSummaryView",props:{items:{}},setup(w){const b=w;return(Q,r)=>{const R=l("RouteTitle"),y=l("XAction"),z=l("XSelect"),h=l("XLayout"),F=l("DataLoader"),S=l("AppView"),V=l("DataCollection"),E=l("RouteView");return d(),i(E,{name:"hostname-generator-summary-view",params:{name:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:String}},{default:o(({route:e,t:c,can:M,uri:B})=>[m(V,{items:b.items,predicate:s=>s.id===e.params.name},{item:o(({item:s})=>[m(S,null,{title:o(()=>[u("h2",null,[m(y,{to:{name:"hostname-generator-detail-view",params:{name:e.params.name}}},{default:o(()=>[m(R,{title:c("hostname-generators.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:o(()=>[r[7]||(r[7]=t()),m(h,{type:"stack"},{default:o(()=>[u("header",null,[m(h,{type:"separated",size:"max"},{default:o(()=>[u("h3",null,p(c("hostname-generators.routes.item.config")),1),r[0]||(r[0]=t()),(d(),C(X,null,k([["structured","universal","k8s"]],n=>u("div",{key:typeof n},[m(z,{label:c("hostname-generators.routes.item.format"),selected:e.params.format,onChange:a=>{e.update({format:a})},onVnodeBeforeMount:a=>{var _;return((_=a==null?void 0:a.props)==null?void 0:_.selected)&&n.includes(a.props.selected)&&a.props.selected!==e.params.format&&e.update({format:a.props.selected})}},q({_:2},[k(n,a=>({name:`${a}-option`,fn:o(()=>[t(p(c(`hostname-generators.routes.item.formats.${a}`)),1)])}))]),1032,["label","selected","onChange","onVnodeBeforeMount"])])),64))]),_:2},1024)]),r[6]||(r[6]=t()),e.params.format==="structured"?(d(),C("div",N,[s.namespace.length>0?(d(),i(g,{key:0,layout:"horizontal"},{title:o(()=>[t(p(c("hostname-generators.common.namespace")),1)]),body:o(()=>[t(p(s.namespace),1)]),_:2},1024)):f("",!0),r[4]||(r[4]=t()),M("use zones")&&s.zone?(d(),i(g,{key:1,layout:"horizontal"},{title:o(()=>[t(p(c("hostname-generators.common.zone")),1)]),body:o(()=>[m(y,{to:{name:"zone-cp-detail-view",params:{zone:s.zone}}},{default:o(()=>[t(p(s.zone),1)]),_:2},1032,["to"])]),_:2},1024)):f("",!0),r[5]||(r[5]=t()),s.spec.template?(d(),i(g,{key:2,layout:"horizontal"},{title:o(()=>[t(p(c("hostname-generators.common.template")),1)]),body:o(()=>[t(p(s.spec.template),1)]),_:2},1024)):f("",!0)])):e.params.format==="universal"?(d(),i(x,{key:1,"data-testid":"codeblock-yaml-universal",language:"yaml",resource:s.$raw,"show-k8s-copy-button":!1,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(d(),i(F,{key:2,src:B(L(A),"/hostname-generators/:name/as/kubernetes",{name:e.params.name})},{default:o(({data:n})=>[m(x,{"data-testid":"codeblock-yaml-k8s",language:"yaml",resource:n,"show-k8s-copy-button":!1,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{G as default};
