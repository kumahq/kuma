import{d as C,e as o,o as h,m as x,w as s,a,b as t,l as V,aA as E}from"./index-Bxa7bIor.js";import{C as R}from"./CodeBlock-CDtZQtk8.js";const A=C({__name:"ZoneIngressXdsConfigView",setup(w){return(k,y)=>{const d=o("RouteTitle"),c=o("KCheckbox"),r=o("XAction"),i=o("DataLoader"),l=o("KCard"),p=o("AppView"),m=o("RouteView");return h(),x(m,{name:"zone-ingress-xds-config-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,includeEds:!1}},{default:s(({route:e,t:u,uri:g})=>[a(d,{render:!1,title:u("zone-ingresses.routes.item.navigation.zone-ingress-xds-config-view")},null,8,["title"]),t(),a(p,null,{default:s(()=>[a(l,null,{default:s(()=>[a(i,{src:g(V(E),"/zone-ingresses/:name/xds/:endpoints",{name:e.params.zoneIngress,endpoints:String(e.params.includeEds)})},{default:s(({data:_,refresh:f})=>[a(R,{language:"json",code:JSON.stringify(_,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":s(()=>[a(c,{modelValue:e.params.includeEds,"onUpdate:modelValue":n=>e.params.includeEds=n,label:"Include Endpoints"},null,8,["modelValue","onUpdate:modelValue"]),t(),a(r,{action:"refresh",appearance:"primary",onClick:f},{default:s(()=>[t(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{A as default};
