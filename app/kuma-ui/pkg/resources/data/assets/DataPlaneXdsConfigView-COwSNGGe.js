import{d as h,e as a,o as V,p as E,w as t,a as o,b as d,m as R,Y as w}from"./index-D8KcXOkO.js";const b=h({__name:"DataPlaneXdsConfigView",setup(k){return(y,s)=>{const l=a("RouteTitle"),p=a("XCheckbox"),i=a("XAction"),c=a("XCodeBlock"),r=a("DataLoader"),m=a("KCard"),u=a("AppView"),_=a("RouteView");return V(),E(_,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,includeEds:!1}},{default:t(({route:e,t:f,uri:g})=>[o(l,{render:!1,title:f("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"]),s[2]||(s[2]=d()),o(u,null,{default:t(()=>[o(m,null,{default:t(()=>[o(r,{src:g(R(w),"/meshes/:mesh/dataplanes/:name/xds/:endpoints",{mesh:e.params.mesh,name:e.params.dataPlane,endpoints:String(e.params.includeEds)})},{default:t(({data:C,refresh:x})=>[o(c,{language:"json",code:JSON.stringify(C,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":t(()=>[o(p,{modelValue:e.params.includeEds,"onUpdate:modelValue":n=>e.params.includeEds=n,label:"Include Endpoints"},null,8,["modelValue","onUpdate:modelValue"]),s[1]||(s[1]=d()),o(i,{action:"refresh",appearance:"primary",onClick:x},{default:t(()=>s[0]||(s[0]=[d(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{b as default};
